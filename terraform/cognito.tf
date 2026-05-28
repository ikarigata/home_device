# API を保護するためのユーザープール。個人利用なので機能は最小限。
resource "aws_cognito_user_pool" "main" {
  name = "${var.project}-users"

  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_uppercase = false
    require_symbols   = false
  }

  # 管理者がユーザーを作成する運用（自己サインアップは無効）
  admin_create_user_config {
    allow_admin_create_user_only = true
  }
}

# SPA 用アプリクライアント（クライアントシークレットなし）。
resource "aws_cognito_user_pool_client" "spa" {
  name         = "${var.project}-spa"
  user_pool_id = aws_cognito_user_pool.main.id

  generate_secret = false

  explicit_auth_flows = [
    "ALLOW_USER_SRP_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]

  access_token_validity  = 60
  id_token_validity      = 60
  refresh_token_validity = 30

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }
}
