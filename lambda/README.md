# lambda — API バックエンド（Go）

API Gateway (HTTP API) の背後で動く 2 つの Lambda 関数。いずれも `provided.al2023` ランタイム / arm64 でデプロイする。

| 関数 | ルート | 役割 | 主な環境変数 |
| --- | --- | --- | --- |
| `cmd/trigger` | `POST /capture` | IoT Core へ撮影トリガを publish | `IOT_ENDPOINT`, `DEVICE_ID`, `ALLOWED_ORIGIN` |
| `cmd/image` | `GET /image` | 最新画像の署名付き URL を発行 | `S3_BUCKET`, `S3_KEY_PREFIX`, `DEVICE_ID`, `URL_TTL_SECONDS`, `ALLOWED_ORIGIN` |

認証は API Gateway の Cognito JWT authorizer で行うため、Lambda 自身は認証を意識しない。

## ビルド

```sh
make build          # dist/trigger.zip と dist/image.zip を生成（arm64）
make clean
```

生成された zip は Terraform (`terraform/lambda.tf`) が参照してデプロイする。
