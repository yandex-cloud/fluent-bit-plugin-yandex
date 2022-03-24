package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fluent/fluent-bit-go/output"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

type pluginImpl struct {
	mu      sync.RWMutex
	printMu sync.Mutex

	destination *logging.Destination

	defaults *logging.LogEntryDefaults

	keys *parseKeys

	client *client
}

func (p *pluginImpl) init(getConfigValue func(string) string) (int, error) {
	*p = pluginImpl{}

	keys, err := getParseKeys(getConfigValue)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.keys = keys

	destination, err := getDestination(getConfigValue)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.destination = destination

	entryDefaults, err := getDefaults(getConfigValue)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.defaults = entryDefaults

	client, err := getIngestionClient(getConfigValue)
	if err != nil {
		return output.FLB_ERROR, err
	}
	p.client = client

	return output.FLB_OK, nil
}

func (p *pluginImpl) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*logging.IncomingLogEntry, resourceKeys) {
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

func makeTLSConfig(getConfigValue func(string) string) (*tls.Config, error) {
	const CAFileNameKey = "ca_file"
	CAFileName := getConfigValue(CAFileNameKey)
	fmt.Println("yc-logging: make TLS config")

	if CAFileName != "" {
		fmt.Println("yc-logging: create tls config")
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certs pool %w", err)
		}

		r, err := ioutil.ReadFile(CAFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to get ca_file = %s details: %w", CAFileName, err)
		}
		block, _ := pem.Decode(r)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ca_file = %s details: %w", CAFileName, err)
		}
		caCertPool.AddCert(cert)

		config := &tls.Config{
			RootCAs: caCertPool,
		}

		fmt.Println("yc-logging: tls config successful created")

		return config, nil
	}

	return &tls.Config{}, nil
}
