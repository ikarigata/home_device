// Package iot は AWS IoT Core (data plane) へメッセージを publish する。
package iot

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iotdataplane"
)

type Publisher struct {
	client *iotdataplane.Client
}

// New は IoT data plane クライアントを生成する。
// endpoint は iot:Data-ATS エンドポイント（例: xxxx-ats.iot.ap-northeast-1.amazonaws.com）。
func New(ctx context.Context, region, endpoint string) (*Publisher, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := iotdataplane.NewFromConfig(cfg, func(o *iotdataplane.Options) {
		o.BaseEndpoint = aws.String("https://" + endpoint)
	})
	return &Publisher{client: client}, nil
}

// Publish は指定トピックへ payload を送信する（QoS1相当）。
func (p *Publisher) Publish(ctx context.Context, topic string, payload []byte) error {
	_, err := p.client.Publish(ctx, &iotdataplane.PublishInput{
		Topic:   aws.String(topic),
		Qos:     1,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("publish to %s: %w", topic, err)
	}
	return nil
}
