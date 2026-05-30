// Command images は GET /images を処理する Lambda。
// images/{deviceId}/history/ 以下のオブジェクトを ListObjectsV2 で取得し、
// 撮影時刻(LastModified)つきでフロントへ返す。
//
// レスポンス:
//
//	{
//	  "items": [{ "key": "...", "capturedAt": "2026-05-30T12:34:56Z" }, ...],
//	  "truncated": false
//	}
//
// items は撮影時刻の新しい順。サムネは別途 S3 イベント駆動 Lambda で生成する想定で、
// 現状はキーと時刻のみ返す（画像本体の署名 URL は GET /image?key=... で個別に取得）。
package main

import (
	"context"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/ikarigata/home_device/lambda/internal/httpx"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

var (
	s3client      *s3.Client
	bucket        string
	historyPrefix string
)

func init() {
	if o := os.Getenv("ALLOWED_ORIGIN"); o != "" {
		httpx.AllowOrigin = o
	}
	region := getenv("AWS_REGION", "ap-northeast-1")
	bucket = os.Getenv("S3_BUCKET")
	if bucket == "" {
		log.Fatal("S3_BUCKET is required")
	}
	prefix := getenv("S3_KEY_PREFIX", "images")
	deviceID := getenv("DEVICE_ID", "kitchen-1")
	historyPrefix = strings.Trim(prefix, "/") + "/" + deviceID + "/history/"

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	s3client = s3.NewFromConfig(cfg)
}

type item struct {
	Key        string `json:"key"`
	CapturedAt string `json:"capturedAt"`
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	limit := defaultLimit
	if v := req.QueryStringParameters["limit"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			limit = n
		}
	}

	// 60日上限 + 個人用途では総数が高々数千なので、全件取得 → 降順ソート → 上位 limit の単純実装。
	// 物量が増えてきたら、キーに逆ソート可能なタイムスタンプを入れる等で最適化する。
	var all []s3types.Object
	var token *string
	for {
		out, err := s3client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(historyPrefix),
			ContinuationToken: token,
		})
		if err != nil {
			log.Printf("list error: %v", err)
			return httpx.Error(500, "failed to list images"), nil
		}
		all = append(all, out.Contents...)
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}

	// キーは history/YYYY/MM/DD/HHMMSS.jpg で文字列降順 = 時刻降順。
	sort.Slice(all, func(i, j int) bool {
		return aws.ToString(all[i].Key) > aws.ToString(all[j].Key)
	})

	truncated := false
	if len(all) > limit {
		all = all[:limit]
		truncated = true
	}

	items := make([]item, 0, len(all))
	for _, obj := range all {
		items = append(items, item{
			Key:        aws.ToString(obj.Key),
			CapturedAt: aws.ToTime(obj.LastModified).UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return httpx.JSON(200, map[string]any{
		"items":     items,
		"truncated": truncated,
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
