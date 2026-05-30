#!/usr/bin/env bash
# home_device のフルデプロイスクリプト。
# 順序:
#   1. Lambda zip をビルド（terraform が参照するため先に作る必要あり）
#   2. terraform apply で AWS リソース構築/更新
#   3. terraform output から frontend 用の env を生成
#   4. frontend をビルド
#   5. S3 へ sync（assets は長期キャッシュ、index.html は no-cache）
#   6. CloudFront キャッシュ無効化（index.html のみ）
#
# 前提:
#   - AWS CLI が認証済み（aws sts get-caller-identity が通る）
#   - terraform / go / node / npm / zip がインストール済み
#   - terraform/terraform.tfvars が必要に応じて作成済み

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

step() { printf "\n\033[1;36m==> %s\033[0m\n" "$*"; }

step "[1/6] Lambda zip をビルド"
make -C lambda build

step "[2/6] terraform apply"
terraform -chdir=terraform apply

step "[3/6] terraform output から frontend 用 env を生成"
API_ENDPOINT=$(terraform -chdir=terraform output -raw api_endpoint)
USER_POOL_ID=$(terraform -chdir=terraform output -raw user_pool_id)
CLIENT_ID=$(terraform -chdir=terraform output -raw user_pool_client_id)
FRONTEND_BUCKET=$(terraform -chdir=terraform output -raw frontend_bucket)
CLOUDFRONT_ID=$(terraform -chdir=terraform output -raw cloudfront_distribution_id)
FRONTEND_URL=$(terraform -chdir=terraform output -raw frontend_url)

cat > frontend/.env.production <<EOF
VITE_API_ENDPOINT=$API_ENDPOINT
VITE_USER_POOL_ID=$USER_POOL_ID
VITE_CLIENT_ID=$CLIENT_ID
EOF

step "[4/6] frontend をビルド"
(cd frontend && npm install --silent && npm run build)

step "[5/6] S3 へ sync (bucket: $FRONTEND_BUCKET)"
# Vite が dist/assets/ に出すハッシュ付きファイルは内容変われば名前も変わるので長期キャッシュOK
aws s3 sync frontend/dist "s3://$FRONTEND_BUCKET/" \
    --delete \
    --exclude "index.html" \
    --cache-control "public, max-age=31536000, immutable"

# index.html はキャッシュさせない（次回デプロイ即反映）
aws s3 cp frontend/dist/index.html "s3://$FRONTEND_BUCKET/index.html" \
    --cache-control "no-cache, no-store, must-revalidate"

step "[6/6] CloudFront キャッシュ無効化 (distribution: $CLOUDFRONT_ID)"
aws cloudfront create-invalidation \
    --distribution-id "$CLOUDFRONT_ID" \
    --paths "/index.html" >/dev/null

printf "\n\033[1;32m==> デプロイ完了\033[0m\n"
printf "  SPA URL: %s\n" "$FRONTEND_URL"
printf "  ※ CloudFront の初回デプロイは反映まで 15 分前後かかります\n"
