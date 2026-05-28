output "api_endpoint" {
  description = "フロントエンドの VITE_API_ENDPOINT に設定する"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "user_pool_id" {
  description = "VITE_USER_POOL_ID"
  value       = aws_cognito_user_pool.main.id
}

output "user_pool_client_id" {
  description = "VITE_CLIENT_ID"
  value       = aws_cognito_user_pool_client.spa.id
}

output "bucket" {
  value = aws_s3_bucket.images.id
}

output "iot_endpoint" {
  description = "ラズパイ / Lambda が接続する IoT データエンドポイント"
  value       = data.aws_iot_endpoint.ats.endpoint_address
}

output "iot_role_alias" {
  description = "デバイスが S3 用一時クレデンシャルを取得する role alias"
  value       = aws_iot_role_alias.device.alias
}

# ラズパイに配置するデバイス証明書一式（state に平文保存されるため取り扱い注意）。
# 取り出し例: terraform output -raw device_private_key > device.pem.key
output "device_certificate_pem" {
  description = "デバイス証明書 (.pem.crt)"
  value       = aws_iot_certificate.device.certificate_pem
  sensitive   = true
}

output "device_private_key" {
  description = "デバイス秘密鍵 (.pem.key)"
  value       = aws_iot_certificate.device.private_key
  sensitive   = true
}
