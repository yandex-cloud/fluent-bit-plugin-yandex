package yclient

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

	client2 "github.com/yandex-cloud/fluent-bit-plugin-yandex/client"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/config"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/model"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

type client struct {
	mu     sync.RWMutex
	writer logging.LogIngestionServiceClient

	initTime time.Time
}

func (c *client) Write(ctx context.Context, req *model.WriteRequest, opts ...grpc.CallOption) (map[int64]*status.Status, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	in := loggingWriteRequest(req)
	res, err := c.writer.Write(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	return res.GetErrors(), nil
}

func (c *client) Init(authorization string, endpoint string, CAFileName string) error {
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
		grpc.WithUserAgent(`fluent-bit-plugin-yandex/`+config.PluginVersion+`; fluent-bit/`+config.FluentBitVersion),
	)
	if err != nil {
		return fmt.Errorf("error creating sdk: %s", err.Error())
	}
	c.writer = sdk.LogIngestion().LogIngestion()
	c.initTime = time.Now()
	return nil
}

func New(authorization string, endpoint string, CAFileName string) (client2.Client, error) {
	c := new(client)
	return c, c.Init(authorization, endpoint, CAFileName)
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

		conf := &tls.Config{
			RootCAs: caCertPool,
		}

		fmt.Println("yc-logging: tls config successful created")

		return conf, nil
	}

	return &tls.Config{}, nil
}
