package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

type Client interface {
	Write(ctx context.Context, in *logging.WriteRequest, opts ...grpc.CallOption) (*logging.WriteResponse, error)
	Init(authorization string, endpoint string, CAFileName string) error
}
