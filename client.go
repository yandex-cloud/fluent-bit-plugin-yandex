package main

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc"
)

type client struct {
	logging.LogIngestionServiceClient
	reinit func()
}

var (
	PluginVersion    string
	FluentBitVersion string
)

func getIngestionClient(plugin unsafe.Pointer) (*client, error) {
	const (
		keyAuthorization = "authorization"
		keyEndpoint      = "endpoint"
		defaultEndpoint  = "api.cloud.yandex.net:443"
	)

	reinitFunc := func() {}

	authorization := getConfigKey(plugin, keyAuthorization)
	if authorization == "" {
		return nil, fmt.Errorf("authorization missing")
	}

	credentials, err := makeCredentials(authorization)
	if err != nil {
		return nil, err
	}
	if r, ok := credentials.(*refreshableCredentials); ok {
		reinitFunc = r.refresh
	}

	endpoint := getConfigKey(plugin, keyEndpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	tlsConfig, err := makeTLSConfig(plugin)
	if err != nil {
		return nil, fmt.Errorf("error creating tls config: %s", err.Error())
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
		return nil, fmt.Errorf("error creating sdk: %s", err.Error())
	}
	return &client{
		LogIngestionServiceClient: sdk.LogIngestion().LogIngestion(),
		reinit:                    reinitFunc,
	}, nil
}
