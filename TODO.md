# TODO

Terraform レビュー（2026-05-30）で挙がった改善点。

## 要修正

- [x] **デバイス MQTT ポリシーから `iot:Publish` を外す** — `terraform/iot.tf`
  ラズパイは subscribe 側で、publish は Lambda(trigger) が IAM 経由で行う。`Receive` + `Subscribe` のみで十分。最小権限の徹底。

- [x] **`allowed_origin` のデフォルトを `"*"` から絞る** — `terraform/frontend.tf`, `terraform/variables.tf`, `terraform/apigateway.tf`, `terraform/lambda.tf`
  CloudFront+S3 のフロント配信インフラを Terraform で構築し、CORS 許可オリジンを CloudFront ドメインに自動 wire。dev 用 localhost は `extra_allowed_origins` で追加可能。

- [ ] **デバイス用 IAM ロールの trust policy に Source 条件を追加** — `terraform/iam.tf:5-14`
  `credentials.iot.amazonaws.com` の AssumeRole trust に `aws:SourceArn = <role alias ARN>` 条件を入れて Confused Deputy 対策。個人用途では低リスクだが堅くなる。

- [x] **Lambda zip の事前ビルド依存を解消（または手順を明示）** — `deploy.sh`
  Lambda ビルド → terraform apply → frontend ビルド → S3 sync → CloudFront 無効化を順に実行する `deploy.sh` を用意。`./deploy.sh` 一発でデプロイ完結。

## 任意（個人利用なら見送り可）

- [ ] API Gateway stage のスロットリング / アクセスログ設定 — `terraform/apigateway.tf:74-78`
- [ ] Lambda の DLQ・CloudWatch アラーム
- [ ] Terraform state を S3 backend に移行（`terraform/versions.tf:11-16` にコメント済み）
