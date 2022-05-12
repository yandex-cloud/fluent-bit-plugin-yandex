package main

import (
	"fmt"
	"unsafe"

	ycsdk "github.com/yandex-cloud/go-sdk"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/yclient"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/config"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/metadata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"C"

	plugin2 "github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/plugin"

	"github.com/fluent/fluent-bit-go/output"
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	fmt.Println("yc-logging: registering")
	return output.FLBPluginRegister(def, "yc-logging", "Yandex Cloud Logging output")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	fmt.Println("yc-logging: init")

	getConfigValue := func(key string) string {
		return config.GetKey(plugin, key)
	}
	metadataProvider := metadata.NewCachingProvider(ycsdk.InstanceMetadataAddr)

	destination, err := config.GetDestination(getConfigValue, metadataProvider)
	if err != nil {
		return output.FLB_ERROR
	}
	defaults, err := config.GetDefaults(getConfigValue, metadataProvider)
	if err != nil {
		return output.FLB_ERROR
	}

	authorization, err := config.GetAuthorization(getConfigValue, metadataProvider)
	if err != nil {
		return output.FLB_ERROR
	}
	endpoint := config.GetEndpoint(getConfigValue)
	CAFileName := config.GetCAFileName(getConfigValue)

	ingestionClient, err := yclient.New(destination, defaults, authorization, endpoint, CAFileName)
	if err != nil {
		return output.FLB_ERROR
	}

	impl, err := plugin2.New(getConfigValue, metadataProvider, ingestionClient)
	if err != nil {
		fmt.Printf("yc-logging: init err: %s\n", err.Error())
		return output.FLB_ERROR
	}

	output.FLBPluginSetContext(plugin, impl)
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	tagStr := C.GoString(tag)

	plugin := output.FLBPluginGetContext(ctx).(*plugin2.Plugin)

	dec := output.NewDecoder(data, int(length))
	provider := func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
		return output.GetRecord(dec)
	}
	resourceToEntries := plugin.Transform(provider, tagStr)

	results, resCount := plugin.WriteAll(resourceToEntries)

	for i := 0; i < resCount; i++ {
		err := <-results
		if err == nil {
			continue
		}

		code := status.Code(err)
		switch code {
		case codes.PermissionDenied:
			// kick client reinit
			fmt.Printf("yc-logging: reinit on write error %s: %q\n", code.String(), err.Error())
			if initErr := plugin.InitClient(); initErr != nil {
				fmt.Printf("yc-logging: reinit failed: %q\n", initErr.Error())
			} else {
				fmt.Printf("yc-logging: reinit succeded\n")
			}
			return output.FLB_RETRY
		case codes.ResourceExhausted, codes.FailedPrecondition, codes.Unavailable,
			codes.Canceled, codes.DeadlineExceeded:
			fmt.Printf("yc-logging: write retriable error %s: %q\n", code.String(), err.Error())
			return output.FLB_RETRY
		default:
			fmt.Printf("yc-logging: write failed %s: %q\n", code.String(), err.Error())
			return output.FLB_ERROR
		}
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
