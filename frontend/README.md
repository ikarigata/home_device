# frontend — スマホ向け Web アプリ

React + Vite (TypeScript) の SPA。Cognito でログインし、「撮影する」ボタンで撮影をトリガして最新画像を表示する。

## セットアップ

```sh
npm install
cp .env.example .env   # terraform output の値を設定
npm run dev            # 開発サーバ
npm run build          # dist/ に本番ビルド
```

`.env` に設定する値（`cd ../terraform && terraform output` で取得）:

| 変数 | 対応する output |
| --- | --- |
| `VITE_API_ENDPOINT` | `api_endpoint` |
| `VITE_USER_POOL_ID` | `user_pool_id` |
| `VITE_CLIENT_ID` | `user_pool_client_id` |

## ログインユーザーの作成

ユーザープールは管理者作成のみ（自己サインアップ無効）。AWS CLI で初期ユーザーを作成する例:

```sh
aws cognito-idp admin-create-user --user-pool-id <POOL_ID> --username <name>
aws cognito-idp admin-set-user-password --user-pool-id <POOL_ID> --username <name> --password <pw> --permanent
```

## デプロイ（例）

`dist/` を S3 静的ホスティング + CloudFront に配置。配信オリジンを Terraform の `allowed_origin` に設定して CORS を絞る。
