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
4. **エッジ**: ラズパイに `edge` をデプロイ、IoT デバイス証明書を配置し `CAMERA=fswebcam` で常駐起動。
   - 開発マシンで擬似動作を試す場合は `CAMERA=mock go run ./cmd/agent` でダミー画像のパイプライン検証も可能。

> 現状の準備状況: AWS アカウント / ラズパイ実機 / カメラ（Logicool C270nd）まで手配・接続済み（`fswebcam` での撮影動作も確認済み）。独自ドメインのみ未手配。AWS 側のリソース構築（terraform apply 以降）はこれから。

## 開発環境

`.devcontainer/` に Go / Node / AWS CLI / Terraform / GitHub CLI が同梱。VS Code の Dev Containers で開けばそのまま開発できる。
