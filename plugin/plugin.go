package plugin

import (
	"fmt"
	"time"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/config"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/metadata"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/util"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

type nextRecordProvider func() (ret int, ts interface{}, rec map[interface{}]interface{})

type Plugin struct {
	getConfigValue   func(string) string
	metadataProvider metadata.MetadataProvider

	destination *logging.Destination
	defaults    *logging.LogEntryDefaults
	keys        *parseKeys

	client client.Client
}

func New(getConfigValue func(string) string, metadataProvider metadata.MetadataProvider, ingestionClient client.Client) (*Plugin, error) {
	p := &Plugin{
		getConfigValue:   getConfigValue,
		metadataProvider: metadataProvider,
	}

	keys := getParseKeys(getConfigValue, metadataProvider)
	p.keys = keys

	destination, err := getDestination(getConfigValue, metadataProvider)
	if err != nil {
		return nil, err
	}
	p.destination = destination

	entryDefaults, err := getDefaults(getConfigValue, metadataProvider)
	if err != nil {
		return nil, err
	}
	p.defaults = entryDefaults

	p.client = ingestionClient

	return p, nil
}

func (p *Plugin) InitClient() error {
	authorization, err := config.GetAuthorization(p.getConfigValue, p.metadataProvider)
	if err != nil {
		return err
	}
	endpoint := config.GetEndpoint(p.getConfigValue)
	CAFileName := config.GetCAFileName(p.getConfigValue)
	return p.client.Init(authorization, endpoint, CAFileName)
}

func (p *Plugin) Transform(provider nextRecordProvider, tag string) map[Resource][]*logging.IncomingLogEntry {
	resourceToEntries := make(map[Resource][]*logging.IncomingLogEntry)

	for {
		ret, ts, record := provider()
		if ret != 0 {
			break
		}

		entry, res, err := p.entry(util.ToTime(ts), record, tag)
		if err != nil {
			fmt.Printf("yc-logging: could not write entry %v because of error: %s\n", record, err.Error())
			continue
		}
		entries, ok := resourceToEntries[res]
		if ok {
			entries = append(entries, entry)
		} else {
			entries = []*logging.IncomingLogEntry{entry}
		}
		resourceToEntries[res] = entries
	}

	return resourceToEntries
}

func (p *Plugin) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*logging.IncomingLogEntry, Resource, error) {
	return p.keys.entry(ts, record, tag)
}
