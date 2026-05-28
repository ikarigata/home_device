data "aws_caller_identity" "current" {}

# ラズパイ / Lambda が接続する IoT Core のデータエンドポイント
data "aws_iot_endpoint" "ats" {
  endpoint_type = "iot:Data-ATS"
}

locals {
  bucket_name   = "${var.project}-images-${data.aws_caller_identity.current.account_id}"
  capture_topic = "home_device/${var.device_id}/capture"
  object_key    = "${var.s3_key_prefix}/${var.device_id}/latest.jpg"
}
