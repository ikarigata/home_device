// Command image は GET /image を処理する Lambda。
// 最新画像 (images/{deviceId}/latest.jpg) の署名付き GET URL を発行して返す。
package main

import (
	"context"
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
	presigner *s3url.Presigner
	objectKey string
	urlTTL    time.Duration
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
	objectKey = strings.Trim(prefix, "/") + "/" + deviceID + "/latest.jpg"

	ttlSec, _ := strconv.Atoi(getenv("URL_TTL_SECONDS", "300"))
	urlTTL = time.Duration(ttlSec) * time.Second

	var err error
	presigner, err = s3url.New(context.Background(), region, bucket)
	if err != nil {
		log.Fatalf("init presigner: %v", err)
	}
}

func handler(ctx context.Context, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	url, err := presigner.GetURL(ctx, objectKey, urlTTL)
	if err != nil {
		log.Printf("presign error: %v", err)
		return httpx.Error(500, "failed to create image url"), nil
	}
	return httpx.JSON(200, map[string]any{
		"url":       url,
		"expiresIn": int(urlTTL.Seconds()),
	}), nil
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
