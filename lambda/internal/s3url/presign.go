// Package s3url は S3 オブジェクトの署名付き GET URL を発行する。
package s3url

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ErrNotFound はオブジェクトが存在しない場合に LastModified が返す sentinel エラー。
var ErrNotFound = errors.New("object not found")

type Presigner struct {
	client  *s3.Client
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
		client:  client,
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

// LastModified は key の S3 オブジェクトの最終更新時刻を返す。
// オブジェクトが存在しない場合は ErrNotFound を返す（撮影完了ポーリング中に利用）。
func (p *Presigner) LastModified(ctx context.Context, key string) (time.Time, error) {
	head, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nf *s3types.NotFound
		if errors.As(err, &nf) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("head %s: %w", key, err)
	}
	if head.LastModified == nil {
		return time.Time{}, fmt.Errorf("head %s: missing LastModified", key)
	}
	return *head.LastModified, nil
}
