package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yandex-cloud/go-sdk/iamkey"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type Client struct {
	mu     sync.RWMutex
	writer logging.LogIngestionServiceClient

	initTime time.Time
	Init     func() error
}

var _ logging.LogIngestionServiceClient = (*Client)(nil)

func (c *Client) Write(ctx context.Context, in *logging.WriteRequest, opts ...grpc.CallOption) (*logging.WriteResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.writer.Write(ctx, in, opts...)
}

func New(authorization string, endpoint string, CAFileName string) (*Client, error) {
	c := new(Client)
	c.Init = clientInit(c, authorization, endpoint, CAFileName)
	return c, c.Init()
}

var (
	PluginVersion    string
	FluentBitVersion string
)

func clientInit(c *Client, authorization string, endpoint string, CAFileName string) func() error {
	return func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		const initBackoff = 30 * time.Second
		passed := time.Since(c.initTime)
		if passed < initBackoff {
			return fmt.Errorf("%s since last client init haven't passed, only %s", initBackoff, passed)
		}

		credentials, err := makeCredentials(authorization)
		if err != nil {
			return err
		}

		tlsConfig, err := makeTLSConfig(CAFileName)
		if err != nil {
			return fmt.Errorf("error creating tls config: %s", err.Error())
		}

		sdk, err := ycsdk.Build(context.Background(),
			ycsdk.Config{
				Credentials: credentials,
				Endpoint:    endpoint,
				TLSConfig:   tlsConfig,
			},
			grpc.WithUserAgent(`fluent-bit-plugin-yandex/`+PluginVersion+`; fluent-bit/`+FluentBitVersion),
		)
		if err != nil {
			return fmt.Errorf("error creating sdk: %s", err.Error())
		}
		c.writer = sdk.LogIngestion().LogIngestion()
		c.initTime = time.Now()
		return nil
	}
}

func makeCredentials(authorization string) (ycsdk.Credentials, error) {
	const (
		instanceSaAuth   = "instance-service-account"
		tokenAuth        = "iam-token"
		iamKeyAuthPrefix = "iam-key-file:"
	)
	switch auth := strings.TrimSpace(authorization); auth {
	case instanceSaAuth:
		return ycsdk.InstanceServiceAccount(), nil
	case tokenAuth:
		token, ok := os.LookupEnv("YC_TOKEN")
		if !ok {
			return nil, errors.New(`environment variable "YC_TOKEN" not set, required for authorization=iam-token`)
		}
		return ycsdk.NewIAMTokenCredentials(token), nil
	default:
		if !strings.HasPrefix(auth, iamKeyAuthPrefix) {
			return nil, fmt.Errorf("unsupported authorization parameter %s", auth)
		}
		fileName := strings.TrimSpace(auth[len(iamKeyAuthPrefix):])
		key, err := iamkey.ReadFromJSONFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account key file %s", fileName)
		}
		return ycsdk.ServiceAccountKey(key)
	}
}

func makeTLSConfig(CAFileName string) (*tls.Config, error) {
	fmt.Println("yc-logging: make TLS config")

	if CAFileName != "" {
		fmt.Println("yc-logging: create tls config")
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certs pool %w", err)
		}

		r, err := ioutil.ReadFile(CAFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to get ca_file = %s details: %w", CAFileName, err)
		}
		block, _ := pem.Decode(r)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ca_file = %s details: %w", CAFileName, err)
		}
		caCertPool.AddCert(cert)

		config := &tls.Config{
			RootCAs: caCertPool,
		}

		fmt.Println("yc-logging: tls config successful created")

		return config, nil
	}

	return &tls.Config{}, nil
}
