package test

import (
	"context"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
)

type Client struct{}

func (c *Client) Write(ctx context.Context, in *model.WriteRequest, opts ...grpc.CallOption) (map[int64]*status.Status, error) {
	_ = ctx
	_ = in
	_ = opts
	return nil, nil
}

func (c *Client) Init(authorization string, endpoint string, CAFileName string) error {
	_ = authorization
	_ = endpoint
	_ = CAFileName
	return nil
}
