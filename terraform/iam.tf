#############################################
# デバイス(ラズパイ)用 IAM: S3 への最小権限
# IoT credentials provider 経由で一時クレデンシャルを取得して使う
#############################################
locals {
  device_role_alias_name = "${var.project}-device-s3"
  device_role_alias_arn  = "${local.iot_arn_prefix}:rolealias/${local.device_role_alias_name}"
}

# Confused Deputy 対策: この role を AssumeRole できるのは
# 当該 role alias 経由の credentials provider 呼び出しに限定する。
data "aws_iam_policy_document" "device_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["credentials.iot.amazonaws.com"]
    }
    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [local.device_role_alias_arn]
    }
  }
}

resource "aws_iam_role" "device" {
  name               = "${var.project}-device-role"
  assume_role_policy = data.aws_iam_policy_document.device_assume.json
}

resource "aws_iam_role_policy" "device_s3" {
  name = "s3-put-images"
  role = aws_iam_role.device.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:PutObject"
      Resource = "${aws_s3_bucket.images.arn}/${var.s3_key_prefix}/${var.device_id}/*"
    }]
  })
}

resource "aws_iot_role_alias" "device" {
  alias    = local.device_role_alias_name
  role_arn = aws_iam_role.device.arn
}

#############################################
# Lambda 実行ロール
#############################################
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# trigger Lambda: IoT publish のみ
resource "aws_iam_role" "lambda_trigger" {
  name               = "${var.project}-lambda-trigger"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_iam_role_policy_attachment" "trigger_logs" {
  role       = aws_iam_role.lambda_trigger.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "trigger_iot" {
  name = "iot-publish"
  role = aws_iam_role.lambda_trigger.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "iot:Publish"
      Resource = "${local.iot_arn_prefix}:topic/${local.capture_topic}"
    }]
  })
}

# image Lambda: latest と履歴の GetObject（署名付き URL は署名者の権限を継承する）
resource "aws_iam_role" "lambda_image" {
  name               = "${var.project}-lambda-image"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_iam_role_policy_attachment" "image_logs" {
  role       = aws_iam_role.lambda_image.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "image_s3" {
  name = "s3-get-image"
  role = aws_iam_role.lambda_image.id

  # ListBucket がないと未生成オブジェクトの HeadObject が 404 ではなく 403 で返る
  # （S3 はオブジェクト存在の有無を漏らさないため）。撮影完了ポーリング中の 404 判定に必要。
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.images.arn}/${var.s3_key_prefix}/${var.device_id}/*"
      },
      {
        Effect   = "Allow"
        Action   = "s3:ListBucket"
        Resource = aws_s3_bucket.images.arn
        Condition = {
          StringLike = {
            "s3:prefix" = ["${var.s3_key_prefix}/${var.device_id}/*"]
          }
        }
      },
    ]
  })
}

# images Lambda: 履歴一覧のための ListBucket（自デバイスの history/ プレフィックス限定）
resource "aws_iam_role" "lambda_images" {
  name               = "${var.project}-lambda-images"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_iam_role_policy_attachment" "images_logs" {
  role       = aws_iam_role.lambda_images.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "images_list" {
  name = "s3-list-history"
  role = aws_iam_role.lambda_images.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:ListBucket"
      Resource = aws_s3_bucket.images.arn
      Condition = {
        StringLike = {
          "s3:prefix" = ["${var.s3_key_prefix}/${var.device_id}/history/*"]
        }
      }
    }]
  })
}
