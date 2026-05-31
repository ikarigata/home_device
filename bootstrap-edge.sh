#!/usr/bin/env bash
# ラズパイの初回セットアップ + CSR ベースのデバイス証明書プロビジョニング。
#
# CSR 方式: 秘密鍵はラズパイ上で生成され、決して外に出ない。
# AWS には公開情報である CSR のみを送り、署名済み cert を受け取る。
# よって terraform state にも秘密鍵は入らない。
#
# 前提:
#   - AWS CLI 認証済み（aws sts get-caller-identity が通る）
#   - terraform / go / ssh / scp が使える
#   - ~/.ssh/config に "Host ika" 定義済み（または SSH_HOST 環境変数）
#   - ラズパイ側で sudo がパスワードなしで使える
#   - terraform/terraform.tfvars が必要に応じて作成済み
#
# 何が起きるか（idempotent: 何度実行しても安全）:
#   1. ラズパイに home_device ユーザー / /etc/home_device/certs を準備
#   2. ラズパイ上で秘密鍵 (device.pem.key) を生成（既存なら再利用）
#   3. その鍵から CSR を作成 → 開発機の terraform/device.csr に取得
#   4. Lambda zip ビルド + terraform apply (CSR を AWS が署名 → cert 発行)
#      ※ frontend / CloudFront のデプロイは別途 ./deploy.sh で
#   5. 署名済み cert (device.pem.crt) を取得 → ラズパイへ配置
#   6. Amazon Root CA を配置
#   7. terraform output から agent.env を生成 → ラズパイへ配置
#   8. agent バイナリを arm64 でビルドし scp + install
#   9. systemd unit を配置して enable --now
#
# 2回目以降の運用:
#   - サーバー側（Lambda / TF / frontend）の更新: ./deploy.sh
#   - エッジバイナリの差し替え: ./deploy-edge.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

# terraform/aws CLI 用プロファイル。dev ユーザーには IAM 作成権限がないため必須。
export AWS_PROFILE="${AWS_PROFILE:-terraform}"

SSH_HOST="${SSH_HOST:-ika}"
GOARCH="${GOARCH:-arm64}"
GOARM="${GOARM:-}"

DEVICE_ID="${DEVICE_ID:-kitchen-1}" # CSR の CN に使う。terraform.tfvars と必ず一致させる

CSR_LOCAL="$REPO_ROOT/terraform/device.csr"
CERT_DIR_REMOTE="/etc/home_device/certs"
KEY_REMOTE="$CERT_DIR_REMOTE/device.pem.key"
CRT_REMOTE="$CERT_DIR_REMOTE/device.pem.crt"
CA_REMOTE="$CERT_DIR_REMOTE/AmazonRootCA1.pem"
ENV_REMOTE="/etc/home_device/agent.env"
UNIT_REMOTE="/etc/systemd/system/home_device.service"

step() { printf "\n\033[1;36m==> %s\033[0m\n" "$*"; }

step "[1/9] ラズパイにユーザー・ディレクトリを準備"
ssh "$SSH_HOST" 'bash -s' <<EOF
set -euo pipefail
if ! id home_device >/dev/null 2>&1; then
    sudo useradd --system --no-create-home --shell /usr/sbin/nologin -G video home_device
    echo "  created user: home_device"
else
    echo "  user home_device は既存（スキップ）"
fi
sudo install -m 0750 -o root -g home_device -d $CERT_DIR_REMOTE
sudo install -m 0750 -o root -g home_device -d /etc/home_device
EOF

step "[2/9] 秘密鍵を生成（既存ならスキップ）"
ssh "$SSH_HOST" 'bash -s' <<EOF
set -euo pipefail
if sudo test -f $KEY_REMOTE; then
    echo "  既存の秘密鍵を検出: $KEY_REMOTE（スキップ）"
else
    sudo openssl genrsa -out $KEY_REMOTE 2048
    sudo chown root:home_device $KEY_REMOTE
    sudo chmod 0640 $KEY_REMOTE
    echo "  生成: $KEY_REMOTE"
fi
EOF

step "[3/9] CSR を作成 → 開発機に取得"
# CSR は公開情報なので毎回作り直して OK（鍵さえ変わらなければ同じ署名対象）
ssh "$SSH_HOST" "sudo openssl req -new -key $KEY_REMOTE -subj '/CN=$DEVICE_ID'" > "$CSR_LOCAL"
echo "  保存: $CSR_LOCAL"

step "[4/9] Lambda zip ビルド + terraform apply"
make -C lambda build
terraform -chdir=terraform apply -auto-approve

step "[5/9] 署名済み cert をラズパイへ配置"
terraform -chdir=terraform output -raw device_certificate_pem \
    | ssh "$SSH_HOST" "sudo tee $CRT_REMOTE >/dev/null && sudo chown root:home_device $CRT_REMOTE && sudo chmod 0640 $CRT_REMOTE"
echo "  配置: $CRT_REMOTE"

step "[6/9] Amazon Root CA を配置"
ssh "$SSH_HOST" 'bash -s' <<EOF
set -euo pipefail
if sudo test -f $CA_REMOTE; then
    echo "  既存 CA を検出（スキップ）"
else
    curl -sSL -o /tmp/AmazonRootCA1.pem https://www.amazontrust.com/repository/AmazonRootCA1.pem
    sudo install -m 0640 -o root -g home_device /tmp/AmazonRootCA1.pem $CA_REMOTE
    rm -f /tmp/AmazonRootCA1.pem
    echo "  配置: $CA_REMOTE"
fi
EOF

step "[7/9] terraform output から agent.env を生成 → 配置"
S3_BUCKET=$(terraform -chdir=terraform output -raw bucket)
IOT_ENDPOINT=$(terraform -chdir=terraform output -raw iot_endpoint)
IOT_CRED_ENDPOINT=$(terraform -chdir=terraform output -raw iot_cred_endpoint)
IOT_ROLE_ALIAS=$(terraform -chdir=terraform output -raw iot_role_alias)
DEVICE_THING_NAME=$(terraform -chdir=terraform output -raw device_thing_name)
AWS_REGION=$(terraform -chdir=terraform output -raw region)

ENV_BODY=$(
    cat <<EOF
DEVICE_ID=$DEVICE_ID
CAMERA=fswebcam
AWS_REGION=$AWS_REGION
S3_BUCKET=$S3_BUCKET
S3_KEY_PREFIX=images
IOT_ENDPOINT=$IOT_ENDPOINT
IOT_CERT_FILE=$CRT_REMOTE
IOT_KEY_FILE=$KEY_REMOTE
IOT_CA_FILE=$CA_REMOTE
IOT_CRED_ENDPOINT=$IOT_CRED_ENDPOINT
IOT_ROLE_ALIAS=$IOT_ROLE_ALIAS
IOT_THING_NAME=$DEVICE_THING_NAME
EOF
)
printf '%s\n' "$ENV_BODY" \
    | ssh "$SSH_HOST" "sudo tee $ENV_REMOTE >/dev/null && sudo chown root:home_device $ENV_REMOTE && sudo chmod 0640 $ENV_REMOTE"
echo "  配置: $ENV_REMOTE"

step "[8/9] agent バイナリをビルドして配置"
mkdir -p "$REPO_ROOT/edge/dist"
build_env=(GOOS=linux GOARCH="$GOARCH" CGO_ENABLED=0)
if [[ -n "$GOARM" ]]; then build_env+=(GOARM="$GOARM"); fi
(cd "$REPO_ROOT/edge" && env "${build_env[@]}" go build -o dist/agent ./cmd/agent)
scp "$REPO_ROOT/edge/dist/agent" "$SSH_HOST:/tmp/home_device-agent.new"

step "[9/9] systemd unit を配置 → enable --now"
scp "$REPO_ROOT/edge/deploy/home_device.service" "$SSH_HOST:/tmp/home_device.service"
ssh "$SSH_HOST" 'bash -s' <<EOF
set -euo pipefail
sudo install -m 0755 /tmp/home_device-agent.new /usr/local/bin/home_device-agent
sudo install -m 0644 /tmp/home_device.service $UNIT_REMOTE
rm -f /tmp/home_device-agent.new /tmp/home_device.service
sudo systemctl daemon-reload
sudo systemctl enable --now home_device
sleep 1
sudo systemctl status home_device --no-pager -n 10 || true
EOF

printf "\n\033[1;32m==> ブートストラップ完了\033[0m\n"
printf "  以降は ./deploy.sh (サーバー側) / ./deploy-edge.sh (エッジバイナリ) で更新できます。\n"
printf "  ログ追尾: ssh %s 'sudo journalctl -u home_device -f'\n" "$SSH_HOST"
printf "  ※ frontend / CloudFront をまだデプロイしていなければ ./deploy.sh を別途実行してください。\n"
