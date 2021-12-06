package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc"
)

type client struct {
	mu     sync.RWMutex
	writer logging.LogIngestionServiceClient
	init   func() error
}

var _ logging.LogIngestionServiceClient = (*client)(nil)

func (c *client) Write(ctx context.Context, in *logging.WriteRequest, opts ...grpc.CallOption) (*logging.WriteResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.writer.Write(ctx, in, opts...)
}

var (
	PluginVersion    string
	FluentBitVersion string
)

func clientInit(c *client, plugin unsafe.Pointer) func() error {
	var initTime time.Time
	return func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		const initBackoff = 30 * time.Second
		passed := time.Since(initTime)
		if passed < initBackoff {
			return fmt.Errorf("%s since last client init haven't passed, only %s", initBackoff, passed)
		}

		const (
			keyAuthorization = "authorization"
			keyEndpoint      = "endpoint"
			defaultEndpoint  = "api.cloud.yandex.net:443"
		)

		authorization := getConfigKey(plugin, keyAuthorization)
		if authorization == "" {
			return fmt.Errorf("authorization missing")
		}

		credentials, err := makeCredentials(authorization)
		if err != nil {
			return err
		}

		endpoint := getConfigKey(plugin, keyEndpoint)
		if endpoint == "" {
			endpoint = defaultEndpoint
		}

		tlsConfig, err := makeTLSConfig(plugin)
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
		initTime = time.Now()
		return nil
	}
}

func getIngestionClient(plugin unsafe.Pointer) (*client, error) {
	c := new(client)
	c.init = clientInit(c, plugin)
	return c, c.init()
}
