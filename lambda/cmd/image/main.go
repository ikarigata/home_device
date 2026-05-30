// Command image は GET /image を処理する Lambda。
// クエリ key が無ければ最新画像 (images/{deviceId}/latest.jpg) の、
// 指定があればその履歴画像の署名付き GET URL を発行して返す。
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/ikarigata/home_device/lambda/internal/httpx"
	"github.com/ikarigata/home_device/lambda/internal/s3url"
)

var (
	presigner   *s3url.Presigner
	latestKey   string
	deviceScope string // 許可するキーの prefix（例: images/kitchen-1/）
	urlTTL      time.Duration
)

func init() {
	if o := os.Getenv("ALLOWED_ORIGIN"); o != "" {
		httpx.AllowOrigin = o
	}
	region := getenv("AWS_REGION", "ap-northeast-1")
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		log.Fatal("S3_BUCKET is required")
	}
	prefix := getenv("S3_KEY_PREFIX", "images")
	deviceID := getenv("DEVICE_ID", "kitchen-1")
	deviceScope = strings.Trim(prefix, "/") + "/" + deviceID + "/"
	latestKey = deviceScope + "latest.jpg"

	ttlSec, _ := strconv.Atoi(getenv("URL_TTL_SECONDS", "300"))
	urlTTL = time.Duration(ttlSec) * time.Second

	var err error
	presigner, err = s3url.New(context.Background(), region, bucket)
	if err != nil {
		log.Fatalf("init presigner: %v", err)
	}
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	key := req.QueryStringParameters["key"]
	if key == "" {
		key = latestKey
	} else if !isAllowedKey(key) {
		// IDOR 対策: 他デバイス/他バケット配下を覗かれないよう、許可スコープ外は弾く。
		return httpx.Error(400, "invalid key"), nil
	}

	// 撮影完了ポーリングのため LastModified を返す。
	// オブジェクト未生成（初回起動・撮影中）は 404 を返し、フロントはポーリングを続けられる。
	lastModified, err := presigner.LastModified(ctx, key)
	if err != nil {
		if errors.Is(err, s3url.ErrNotFound) {
			return httpx.Error(404, "image not found"), nil
		}
		log.Printf("head error: %v", err)
		return httpx.Error(500, "failed to fetch image metadata"), nil
	}

	url, err := presigner.GetURL(ctx, key, urlTTL)
	if err != nil {
		log.Printf("presign error: %v", err)
		return httpx.Error(500, "failed to create image url"), nil
	}
	return httpx.JSON(200, map[string]any{
		"url":        url,
		"key":        key,
		"capturedAt": lastModified.UTC().Format(time.RFC3339),
		"expiresIn":  int(urlTTL.Seconds()),
	}), nil
}

// isAllowedKey はキーが自デバイスのスコープ内かを検証する。
// ".." を含むパスや先頭スラッシュなど、トラバーサル系も弾く。
func isAllowedKey(key string) bool {
	if !strings.HasPrefix(key, deviceScope) {
		return false
	}
	if strings.Contains(key, "..") {
		return false
	}
	return true
}

func main() {
	lambda.Start(handler)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
