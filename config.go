package main

import (
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

func getConfigKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}
