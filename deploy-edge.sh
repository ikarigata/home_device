#!/usr/bin/env bash
# ラズパイ用エージェントの更新デプロイスクリプト。
# 初回セットアップ（ユーザー作成・証明書配置・systemd unit 登録）は
# edge/README.md の手順を参照（証明書の扱いがあるため自動化していない）。
#
# このスクリプトがやること:
#   1. edge を arm64 向けにクロスコンパイル
#   2. ラズパイ（SSH ホスト: ika）に scp で転送
#   3. sudo install + systemctl restart で反映
#   4. サービス状態を確認
#
# 前提:
#   - ~/.ssh/config に "Host ika" が定義済み
#   - ラズパイ側で home_device.service が既に enable 済み
#   - ラズパイの OS が 64bit Raspberry Pi OS（arm64）
#     32bit の場合は GOARCH=arm GOARM=7 を環境変数で上書き
#
# 例:
#   ./deploy-edge.sh                       # arm64 (デフォルト)
#   GOARCH=arm GOARM=7 ./deploy-edge.sh    # 32bit OS の場合

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT/edge"

SSH_HOST="${SSH_HOST:-ika}"
GOARCH="${GOARCH:-arm64}"
GOARM="${GOARM:-}"

step() { printf "\n\033[1;36m==> %s\033[0m\n" "$*"; }

step "[1/4] edge を Linux/$GOARCH 向けにビルド"
mkdir -p dist
env_vars=(GOOS=linux GOARCH="$GOARCH" CGO_ENABLED=0)
if [[ -n "$GOARM" ]]; then
    env_vars+=(GOARM="$GOARM")
fi
env "${env_vars[@]}" go build -o dist/agent ./cmd/agent
ls -lh dist/agent

step "[2/4] scp で $SSH_HOST に転送"
scp dist/agent "$SSH_HOST:/tmp/home_device-agent.new"

step "[3/4] install + systemctl restart"
ssh "$SSH_HOST" 'bash -s' <<'EOF'
set -euo pipefail
sudo install -m 0755 /tmp/home_device-agent.new /usr/local/bin/home_device-agent
rm -f /tmp/home_device-agent.new
sudo systemctl restart home_device
EOF

step "[4/4] サービス状態を確認"
ssh "$SSH_HOST" 'systemctl is-active home_device && systemctl status home_device --no-pager -n 5'

printf "\n\033[1;32m==> エッジ更新完了\033[0m\n"
printf "  ログ追尾: ssh %s 'sudo journalctl -u home_device -f'\n" "$SSH_HOST"
