package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

type refreshableCredentials struct {
	mu          sync.RWMutex
	initTime    time.Time
	credentials ycsdk.NonExchangeableCredentials
	refreshFunc func() ycsdk.NonExchangeableCredentials
}

var _ ycsdk.NonExchangeableCredentials = (*refreshableCredentials)(nil)

func (r *refreshableCredentials) YandexCloudAPICredentials() {}

func (r *refreshableCredentials) IAMToken(ctx context.Context) (*iam.CreateIamTokenResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.credentials.IAMToken(ctx)
}

func (r *refreshableCredentials) refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()

	const initBackoff = 30 * time.Second
	if time.Since(r.initTime) < initBackoff {
		return
	}
	creds := r.refreshFunc()
	r.credentials = creds
	r.initTime = time.Now()
}

func makeCredentials(authorization string) (credentials ycsdk.Credentials, err error) {
	const (
		instanceSaAuth   = "instance-service-account"
		tokenAuth        = "iam-token"
		tokenKey         = "YC_TOKEN"
		iamKeyAuthPrefix = "iam-key-file:"
	)
	switch auth := strings.TrimSpace(authorization); auth {
	case instanceSaAuth:
		return &refreshableCredentials{
			initTime:    time.Now(),
			credentials: ycsdk.InstanceServiceAccount(),
			refreshFunc: ycsdk.InstanceServiceAccount,
		}, nil
	case tokenAuth:
		token, ok := os.LookupEnv(tokenKey)
		if !ok {
			return nil, errors.New(`environment variable "YC_TOKEN" not set, required for authorization=iam-token`)
		}
		return &refreshableCredentials{
			initTime:    time.Now(),
			credentials: ycsdk.NewIAMTokenCredentials(token),
			refreshFunc: func() ycsdk.NonExchangeableCredentials {
				tok, got := os.LookupEnv(tokenKey)
				if !got {
					fmt.Println(`environment variable "YC_TOKEN" not set, required for authorization=iam-token`)
				}
				return ycsdk.NewIAMTokenCredentials(tok)
			},
		}, nil
	default:
		if !strings.HasPrefix(auth, iamKeyAuthPrefix) {
			return nil, fmt.Errorf("unsupported authorization parameter %s", auth)
		}
		fileName := strings.TrimSpace(auth[len(iamKeyAuthPrefix):])
		key, err := iamkey.ReadFromJSONFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account key file %s", fileName)
		}
		credentials, err = ycsdk.ServiceAccountKey(key)
		return credentials, err
	}
}
