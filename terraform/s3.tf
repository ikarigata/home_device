# 画像保存用の非公開バケット。インターネットから直接アクセスさせず、
# 署名付き URL 経由でのみ配信する。
resource "aws_s3_bucket" "images" {
  bucket = local.bucket_name
}

resource "aws_s3_bucket_public_access_block" "images" {
  bucket = aws_s3_bucket.images.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# 履歴画像のみ60日で自動削除（latest.jpg は pointer なので対象外。
# 撮影が長期間止まっても「最後に撮った絵」だけは残しておく意図）。
resource "aws_s3_bucket_lifecycle_configuration" "images" {
  bucket = aws_s3_bucket.images.id

  rule {
    id     = "expire-history"
    status = "Enabled"

    filter {
      prefix = local.history_prefix
    }

    expiration {
      days = 60
    }
  }
}
