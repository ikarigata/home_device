// Package httpx は API Gateway (HTTP API, payload v2) 向けの共通レスポンスを組み立てる。
package httpx

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// AllowOrigin は CORS の Access-Control-Allow-Origin。
// 環境変数 ALLOWED_ORIGIN で上書きする（lambda 各 main で設定）。
var AllowOrigin = "*"

func headers() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  AllowOrigin,
		"Access-Control-Allow-Headers": "Authorization,Content-Type",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
	}
}

// JSON は任意の値を JSON ボディとして返す。
func JSON(status int, body any) events.APIGatewayV2HTTPResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    headers(),
		Body:       string(b),
	}
}

// Error はエラーメッセージを JSON で返す。
func Error(status int, msg string) events.APIGatewayV2HTTPResponse {
	return JSON(status, map[string]string{"error": msg})
}
