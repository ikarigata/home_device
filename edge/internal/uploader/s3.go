// Package uploader は撮影画像を S3 へアップロードする。
package uploader

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Uploader struct {
	client *s3.Client
	bucket string
}

func New(ctx context.Context, region, bucket string) (*Uploader, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return &Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// PutJPEG は JPEG バイト列を指定キーへ保存する。
func (u *Uploader) PutJPEG(ctx context.Context, key string, data []byte) error {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("put object %s: %w", key, err)
	}
	return nil
}
