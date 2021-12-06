package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

type pluginImpl struct {
	destination *logging.Destination

	resource *logging.LogEntryResource
	defaults *logging.LogEntryDefaults

	keys *parseKeys

	client *client
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

func makeTLSConfig(plugin unsafe.Pointer) (*tls.Config, error) {
	CAFileName := output.FLBPluginConfigKey(plugin, "ca_file")
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
