package main

import (
	"C"
	"context"
	"fmt"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	code, err := impl.init(plugin)
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

	entries := make([]*logging.IncomingLogEntry, 0)
	for {
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}
		entries = append(entries, plugin.entry(toTime(ts), record, tagStr))
	}

	err := plugin.write(context.Background(), entries)

	code := status.Code(err)
	switch code {
	case codes.ResourceExhausted, codes.FailedPrecondition, codes.Unavailable,
		codes.Canceled, codes.DeadlineExceeded:
		fmt.Printf("yc-logging: write retriable error %s: %q\n", code.String(), err.Error())
		return output.FLB_RETRY
	default:
		fmt.Printf("yc-logging: write failed %s: %q\n", code.String(), err.Error())
		return output.FLB_ERROR
	}
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
