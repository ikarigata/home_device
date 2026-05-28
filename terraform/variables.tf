variable "project" {
  description = "リソース名のプレフィックス"
  type        = string
  default     = "home-device"
}

variable "region" {
  description = "デプロイ先リージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "device_id" {
  description = "ラズパイ機器の識別子（MQTTトピック / S3キーに使用）"
  type        = string
  default     = "kitchen-1"
}

variable "allowed_origin" {
  description = "CORS で許可するフロントエンドのオリジン（本番は https://... を指定）"
  type        = string
  default     = "*"
}

variable "s3_key_prefix" {
  description = "画像を保存する S3 キーのプレフィックス"
  type        = string
  default     = "images"
}

variable "url_ttl_seconds" {
  description = "署名付き画像 URL の有効秒数"
  type        = number
  default     = 300
}
