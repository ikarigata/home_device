// Command trigger は POST /capture を処理する Lambda。
// IoT Core の home_device/{deviceId}/capture トピックへ撮影トリガを publish する。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/ikarigata/home_device/lambda/internal/httpx"
	"github.com/ikarigata/home_device/lambda/internal/iot"
)

var (
	publisher *iot.Publisher
	deviceID  string
)

func init() {
	if o := os.Getenv("ALLOWED_ORIGIN"); o != "" {
		httpx.AllowOrigin = o
	}
	deviceID = getenv("DEVICE_ID", "kitchen-1")

	region := getenv("AWS_REGION", "ap-northeast-1")
	endpoint := os.Getenv("IOT_ENDPOINT")
	if endpoint == "" {
		log.Fatal("IOT_ENDPOINT is required")
	}

	var err error
	publisher, err = iot.New(context.Background(), region, endpoint)
	if err != nil {
		log.Fatalf("init iot publisher: %v", err)
	}
}

func handler(ctx context.Context, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	topic := fmt.Sprintf("home_device/%s/capture", deviceID)
	payload, _ := json.Marshal(map[string]string{
		"requestedAt": time.Now().UTC().Format(time.RFC3339),
	})

	if err := publisher.Publish(ctx, topic, payload); err != nil {
		log.Printf("publish error: %v", err)
		return httpx.Error(502, "failed to trigger capture"), nil
	}

	return httpx.JSON(202, map[string]string{
		"status": "capture requested",
		"topic":  topic,
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
