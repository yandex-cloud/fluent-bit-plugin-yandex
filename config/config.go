package config

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/metadata"
)

var (
	PluginVersion    string
	FluentBitVersion string
)

func GetKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func GetAuthorization(getConfigValue func(string) string, metadataProvider metadata.Provider) (string, error) {
	const keyAuthorization = "authorization"

	authorization := getConfigValue(keyAuthorization)
	if authorization == "" {
		return "", fmt.Errorf("authorization missing")
	}

	return metadata.Parse(authorization, metadataProvider), nil
}

func GetEndpoint(getConfigValue func(string) string) string {
	const (
		keyEndpoint     = "endpoint"
		defaultEndpoint = "api.cloud.yandex.net:443"
	)

	endpoint := getConfigValue(keyEndpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	return endpoint
}

func GetCAFileName(getConfigValue func(string) string) string {
	const CAFileNameKey = "ca_file"

	return getConfigValue(CAFileNameKey)
}
