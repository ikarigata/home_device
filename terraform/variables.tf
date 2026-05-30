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

variable "extra_allowed_origins" {
  description = "CloudFront に加えて CORS を許可するオリジン（例: 開発用 [\"http://localhost:5173\"]）"
  type        = list(string)
  default     = []
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

variable "device_csr_path" {
  description = "デバイス証明書発行に使う CSR ファイルのパス。bootstrap-edge.sh がラズパイ上で生成したものを置く。"
  type        = string
  default     = "./device.csr"
}
