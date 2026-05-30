data "aws_caller_identity" "current" {}

# ラズパイ / Lambda が接続する IoT Core のデータエンドポイント
data "aws_iot_endpoint" "ats" {
  endpoint_type = "iot:Data-ATS"
}

# ラズパイが S3 用一時クレデンシャルを取得する credentials provider エンドポイント
data "aws_iot_endpoint" "creds" {
  endpoint_type = "iot:CredentialProvider"
}

locals {
  bucket_name    = "${var.project}-images-${data.aws_caller_identity.current.account_id}"
  capture_topic  = "home_device/${var.device_id}/capture"
  device_prefix  = "${var.s3_key_prefix}/${var.device_id}"
  latest_key     = "${local.device_prefix}/latest.jpg"
  history_prefix = "${local.device_prefix}/history/"
}
