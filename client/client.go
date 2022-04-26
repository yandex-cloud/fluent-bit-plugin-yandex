package client

import (
	"context"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

type Client interface {
	Write(ctx context.Context, in *model.WriteRequest, opts ...grpc.CallOption) (map[int64]*status.Status, error)
	Init(authorization string, endpoint string, CAFileName string) error
}
