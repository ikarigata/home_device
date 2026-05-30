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

| 手順 | コマンド | 何をやるか |
| --- | --- | --- |
| 1. エッジ初回ブートストラップ | `./bootstrap-edge.sh` | ラズパイ側で秘密鍵生成 → CSR を AWS で署名 → cert / agent.env / unit を配置 → サービス起動。Lambda / IoT / S3 などの AWS リソースも `terraform apply` される |
| 2. frontend / CloudFront を出す | `./deploy.sh` | Lambda zip + terraform apply（差分なし）+ frontend ビルド + S3 sync + CloudFront 無効化 |
| 3. Cognito ユーザー作成 | `aws cognito-idp admin-create-user ...` | ログイン用ユーザーを 1 つ作る |

CSR 方式により**秘密鍵はラズパイの外に出ず、tfstate にも入らない**。詳細は [edge/README.md](./edge/README.md)。

開発マシンで擬似動作を試したい場合: `cd edge && CAMERA=mock go run ./cmd/agent`（ダミー画像でパイプラインのみ確認）。

> 現状の準備状況: AWS アカウント / ラズパイ実機 / カメラ（Logicool C270nd）まで手配・接続済み（`fswebcam` での撮影動作も確認済み）。独自ドメインのみ未手配。AWS 側のリソース構築（terraform apply 以降）はこれから。

## 開発環境

`.devcontainer/` に Go / Node / AWS CLI / Terraform / GitHub CLI が同梱。VS Code の Dev Containers で開けばそのまま開発できる。
