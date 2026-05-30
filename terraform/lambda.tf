# `cd ../lambda && make build` で生成した zip を参照する。
locals {
  trigger_zip = "${path.module}/../lambda/dist/trigger.zip"
  image_zip   = "${path.module}/../lambda/dist/image.zip"
  images_zip  = "${path.module}/../lambda/dist/images.zip"
}

resource "aws_cloudwatch_log_group" "trigger" {
  name              = "/aws/lambda/${var.project}-trigger"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "image" {
  name              = "/aws/lambda/${var.project}-image"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "images" {
  name              = "/aws/lambda/${var.project}-images"
  retention_in_days = 14
}

resource "aws_lambda_function" "trigger" {
  function_name = "${var.project}-trigger"
  role          = aws_iam_role.lambda_trigger.arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  filename         = local.trigger_zip
  source_code_hash = filebase64sha256(local.trigger_zip)

  environment {
    variables = {
      DEVICE_ID      = var.device_id
      IOT_ENDPOINT   = data.aws_iot_endpoint.ats.endpoint_address
      ALLOWED_ORIGIN = local.frontend_origin
    }
  }

  depends_on = [aws_cloudwatch_log_group.trigger]
}

resource "aws_lambda_function" "image" {
  function_name = "${var.project}-image"
  role          = aws_iam_role.lambda_image.arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  filename         = local.image_zip
  source_code_hash = filebase64sha256(local.image_zip)

  environment {
    variables = {
      DEVICE_ID       = var.device_id
      S3_BUCKET       = aws_s3_bucket.images.id
      S3_KEY_PREFIX   = var.s3_key_prefix
      URL_TTL_SECONDS = tostring(var.url_ttl_seconds)
      ALLOWED_ORIGIN  = local.frontend_origin
    }
  }

  depends_on = [aws_cloudwatch_log_group.image]
}

resource "aws_lambda_function" "images" {
  function_name = "${var.project}-images"
  role          = aws_iam_role.lambda_images.arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  filename         = local.images_zip
  source_code_hash = filebase64sha256(local.images_zip)

  environment {
    variables = {
      DEVICE_ID      = var.device_id
      S3_BUCKET      = aws_s3_bucket.images.id
      S3_KEY_PREFIX  = var.s3_key_prefix
      ALLOWED_ORIGIN = local.frontend_origin
    }
  }

  depends_on = [aws_cloudwatch_log_group.images]
}
