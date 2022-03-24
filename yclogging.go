package main

import (
	"context"
	"fmt"
	"sync"
	"unsafe"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"C"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	fmt.Println("yc-logging: registering")
	return output.FLBPluginRegister(def, "yc-logging", "Yandex Cloud Logging output")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	fmt.Println("yc-logging: init")

	impl := new(pluginImpl)
	code, err := impl.init(func(key string) string {
		return getConfigKey(plugin, key)
	})
	if err != nil {
		fmt.Printf("yc-logging: init err: %s\n", err.Error())
		return code
	}

	output.FLBPluginSetContext(plugin, impl)
	return code
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	tagStr := C.GoString(tag)

	plugin := output.FLBPluginGetContext(ctx).(*pluginImpl)

	dec := output.NewDecoder(data, int(length))

	resourceToEntries := make(map[resourceKeys][]*logging.IncomingLogEntry)
	for {
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}

		entry, resource := plugin.entry(toTime(ts), record, tagStr)
		entries, ok := resourceToEntries[resource]
		if ok {
			entries = append(entries, entry)
		} else {
			entries = []*logging.IncomingLogEntry{entry}
		}
		resourceToEntries[resource] = entries
	}

	var wg sync.WaitGroup
	resBuffer := len(resourceToEntries)
	results := make(chan error, resBuffer)

	for resource, entries := range resourceToEntries {
		resource := resource.logEntryResource()
		entries := entries

		wg.Add(1)
		go func(res chan error) {
			defer wg.Done()
			err := plugin.write(context.Background(), entries, resource)
			res <- err
		}(results)
	}
	wg.Wait()

	for i := 0; i < resBuffer; i++ {
		err := <-results
		if err == nil {
			continue
		}

		code := status.Code(err)
		switch code {
		case codes.PermissionDenied:
			// kick client reinit
			fmt.Printf("yc-logging: reinit on write error %s: %q\n", code.String(), err.Error())
			if initErr := plugin.client.init(); initErr != nil {
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
