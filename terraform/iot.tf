locals {
  iot_arn_prefix = "arn:aws:iot:${var.region}:${data.aws_caller_identity.current.account_id}"
}

resource "aws_iot_thing" "device" {
  name = "${var.project}-${var.device_id}"
}

# デバイス証明書（CSR 署名方式）。秘密鍵はラズパイ上で生成され外に出ないため、
# state には公開情報の cert しか入らない。
# CSR の用意は bootstrap-edge.sh が担当する。
resource "aws_iot_certificate" "device" {
  active = true
  csr    = file(var.device_csr_path)
}

resource "aws_iot_thing_principal_attachment" "device" {
  thing     = aws_iot_thing.device.name
  principal = aws_iot_certificate.device.arn
}

# 最小権限の MQTT ポリシー:
# - 自分の clientId でのみ接続
# - 撮影トピックの subscribe / receive のみ（publish は Lambda 側が IAM で行う）
# - S3 用一時クレデンシャル取得(credentials provider)のための AssumeRoleWithCertificate
resource "aws_iot_policy" "device" {
  name = "${var.project}-device-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "iot:Connect"
        Resource = "${local.iot_arn_prefix}:client/${var.device_id}"
      },
      {
        Effect   = "Allow"
        Action   = "iot:Receive"
        Resource = "${local.iot_arn_prefix}:topic/${local.capture_topic}"
      },
      {
        Effect   = "Allow"
        Action   = "iot:Subscribe"
        Resource = "${local.iot_arn_prefix}:topicfilter/${local.capture_topic}"
      },
      {
        Effect   = "Allow"
        Action   = "iot:AssumeRoleWithCertificate"
        Resource = aws_iot_role_alias.device.arn
      },
    ]
  })
}

resource "aws_iot_policy_attachment" "device" {
  policy = aws_iot_policy.device.name
  target = aws_iot_certificate.device.arn
}
