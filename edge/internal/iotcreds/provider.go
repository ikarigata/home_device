// Package iotcreds は AWS IoT credentials provider からデバイス証明書(相互TLS)で
// 一時 AWS クレデンシャルを取得する aws.CredentialsProvider を提供する。
//
// 仕組み: HTTPS GET https://<endpoint>/role-aliases/<roleAlias>/credentials を
// デバイス証明書付き(mTLS)で叩くと、role alias に紐づく IAM ロールの一時クレデンシャルが返る。
// これにより、ラズパイに長期 IAM アクセスキーを置かずに S3 へ PutObject できる。
package iotcreds

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Params は credentials provider への接続に必要な値。証明書一式は MQTT と同じものを使う。
type Params struct {
	Endpoint   string // xxxxxxxx.credentials.iot.<region>.amazonaws.com (iot:CredentialProvider)
	RoleAlias  string // terraform output iot_role_alias
	ThingName  string // 任意。指定すると x-amzn-iot-thingname ヘッダを送る
	CertFile   string
	KeyFile    string
	CACertFile string
}

type provider struct {
	url       string
	thingName string
	client    *http.Client
}

// NewProvider は IoT credentials provider から一時クレデンシャルを取得する provider を返す。
// 返り値は aws.NewCredentialsCache でラップしてキャッシュ・自動更新させること。
func NewProvider(p Params) (aws.CredentialsProvider, error) {
	cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load device key pair: %w", err)
	}
	ca, err := os.ReadFile(p.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("parse CA cert")
	}

	return &provider{
		url:       fmt.Sprintf("https://%s/role-aliases/%s/credentials", p.Endpoint, p.RoleAlias),
		thingName: p.ThingName,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      pool,
				},
			},
		},
	}, nil
}

type credentialsResponse struct {
	Credentials struct {
		AccessKeyID     string    `json:"accessKeyId"`
		SecretAccessKey string    `json:"secretAccessKey"`
		SessionToken    string    `json:"sessionToken"`
		Expiration      time.Time `json:"expiration"`
	} `json:"credentials"`
}

func (p *provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return aws.Credentials{}, err
	}
	if p.thingName != "" {
		req.Header.Set("x-amzn-iot-thingname", p.thingName)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("iot credentials request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return aws.Credentials{}, fmt.Errorf("iot credentials provider returned %d: %s", resp.StatusCode, body)
	}

	var cr credentialsResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return aws.Credentials{}, fmt.Errorf("decode credentials: %w", err)
	}

	c := cr.Credentials
	return aws.Credentials{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
		SessionToken:    c.SessionToken,
		Source:          "IoTCredentialsProvider",
		CanExpire:       true,
		Expires:         c.Expiration,
	}, nil
}
