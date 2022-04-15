package test

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/grpc"
)

type Client struct{}

func (c *Client) Write(ctx context.Context, in *logging.WriteRequest, opts ...grpc.CallOption) (*logging.WriteResponse, error) {
	return nil, nil
}
func (c *Client) Init(authorization string, endpoint string, CAFileName string) error {
	return nil
}
