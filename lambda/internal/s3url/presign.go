// Package s3url は S3 オブジェクトの署名付き GET URL を発行する。
package s3url

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Presigner struct {
	presign *s3.PresignClient
	bucket  string
}

func New(ctx context.Context, region, bucket string) (*Presigner, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &Presigner{
		presign: s3.NewPresignClient(client),
		bucket:  bucket,
	}, nil
}

// GetURL は key に対する、ttl の間だけ有効な署名付き GET URL を返す。
func (p *Presigner) GetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := p.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign get %s: %w", key, err)
	}
	return req.URL, nil
}
