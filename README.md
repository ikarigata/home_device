# home_device — 調味料置き場モニター

外出先（スーパーなど）からスマホの Web アプリで自宅の調味料置き場の「今の写真」をオンデマンド撮影・確認し、買い忘れ・二重買いを防ぐサーバーレス IoT システム。

詳しい仕様は [CLAUDE.md](./CLAUDE.md) を参照。

## アーキテクチャ

```
frontend(React) →[JWT]→ API Gateway(HTTP API + Cognito authorizer)
  ├─ POST /capture → lambda(trigger) → IoT Core publish ──┐
  │                                                        ▼
  │                                            edge(ラズパイ) が subscribe
  │                                            → fswebcam 撮影 → S3 PutObject
  └─ GET  /image   → lambda(image) → S3 presigned URL → frontend が表示
```

## ディレクトリ

| ディレクトリ | 役割 | 言語/技術 |
| --- | --- | --- |
| `edge/` | ラズパイ常駐エージェント。MQTT を待受け→撮影→S3 アップロード | Go |
| `lambda/` | API バックエンド（撮影トリガ発行 / 署名付きURL発行） | Go (provided.al2023) |
| `terraform/` | AWS リソース一式の IaC | Terraform |
| `frontend/` | スマホ向け SPA | React + Vite + TypeScript |

## セットアップの流れ（雛形 → 本番）

1. **インフラ構築**: `cd terraform && terraform init && terraform apply`
   - 出力される `api_endpoint` / `user_pool_id` / `client_id` / `bucket` を控える。
2. **Lambda デプロイ**: `cd lambda && make build` で zip を作成し、Terraform 経由で反映。
3. **frontend 設定**: `frontend/.env` に上記 output を設定し `npm run dev` / `npm run build`。
4. **エッジ（実機が来たら）**: ラズパイに `edge` をデプロイ、IoT デバイス証明書を配置し `CAMERA=fswebcam` で常駐起動。
   - 実機が無い間は開発マシンで `CAMERA=mock go run ./cmd/agent` によりダミー画像でパイプラインを検証可能。

> 現状の準備状況: AWS アカウントのみ。ラズパイ実機・カメラ・独自ドメインは未手配。

## 開発環境

`.devcontainer/` に Go / Node / AWS CLI / Terraform / GitHub CLI が同梱。VS Code の Dev Containers で開けばそのまま開発できる。
