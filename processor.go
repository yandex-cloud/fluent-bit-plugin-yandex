package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	ingest "github.com/yandex-cloud/go-sdk/gen/logingestion"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

type pluginImpl struct {
	destination *logging.Destination

	resource *logging.LogEntryResource
	defaults *logging.LogEntryDefaults

	keys *parseKeys

	client *ingest.LogIngestionServiceClient
}

func (p *pluginImpl) init(plugin unsafe.Pointer) (int, error) {
	*p = pluginImpl{
		resource: getResource(plugin),
		keys:     getParseKeys(plugin),
	}

	destination, err := getDestination(plugin)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.destination = destination

	entryDefaults, err := getDefaults(plugin)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.defaults = entryDefaults

	client, err := getIngestionClient(plugin)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.client = client

	return output.FLB_OK, nil
}

func (p *pluginImpl) entry(ts time.Time, record map[interface{}]interface{}, tag string) *logging.IncomingLogEntry {
	return p.keys.entry(ts, record, tag)
}

func makeCredentials(authorization string) (ycsdk.Credentials, error) {
	const (
		instanceSaAuth   = "instance-service-account"
		tokenAuth        = "iam-token"
		iamKeyAuthPrefix = "iam-key-file:"
	)
	switch auth := strings.TrimSpace(authorization); auth {
	case instanceSaAuth:
		return ycsdk.InstanceServiceAccount(), nil
	case tokenAuth:
		token, ok := os.LookupEnv("YC_TOKEN")
		if !ok {
			return nil, errors.New(`environment variable "YC_TOKEN" not set, required for authorization=iam-token`)
		}
		return ycsdk.NewIAMTokenCredentials(token), nil
	default:
		if !strings.HasPrefix(auth, iamKeyAuthPrefix) {
			return nil, fmt.Errorf("unsupported authorization parameter %s", auth)
		}
		fileName := strings.TrimSpace(auth[len(iamKeyAuthPrefix):])
		key, err := iamkey.ReadFromJSONFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read service account key file %s", fileName)
		}
		return ycsdk.ServiceAccountKey(key)
	}
}
