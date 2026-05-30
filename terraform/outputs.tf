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

output "region" {
  description = "デプロイリージョン（agent.env の AWS_REGION）"
  value       = var.region
}

output "frontend_bucket" {
  description = "SPA をアップロードする S3 バケット（aws s3 sync frontend/dist s3://<bucket>/）"
  value       = aws_s3_bucket.frontend.id
}

output "frontend_url" {
  description = "ブラウザでアクセスする SPA の URL"
  value       = local.frontend_origin
}

output "cloudfront_distribution_id" {
  description = "デプロイ後のキャッシュ無効化に使う（aws cloudfront create-invalidation --distribution-id ...）"
  value       = aws_cloudfront_distribution.frontend.id
}

output "iot_endpoint" {
  description = "ラズパイ / Lambda が接続する IoT データエンドポイント"
  value       = data.aws_iot_endpoint.ats.endpoint_address
}

output "iot_cred_endpoint" {
  description = "ラズパイが一時クレデンシャルを取得する IoT credentials provider エンドポイント"
  value       = data.aws_iot_endpoint.creds.endpoint_address
}

output "iot_role_alias" {
  description = "デバイスが S3 用一時クレデンシャルを取得する role alias"
  value       = aws_iot_role_alias.device.alias
}

output "device_thing_name" {
  description = "デバイスの IoT Thing 名（IOT_THING_NAME に設定）"
  value       = aws_iot_thing.device.name
}

# 署名済みデバイス証明書。秘密鍵はラズパイ側で生成・保管されるためここには無い。
output "device_certificate_pem" {
  description = "デバイス証明書 (.pem.crt)。bootstrap-edge.sh がラズパイへ配置する。"
  value       = aws_iot_certificate.device.certificate_pem
  sensitive   = true
}
