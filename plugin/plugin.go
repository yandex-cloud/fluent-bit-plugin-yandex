package plugin

import (
	"fmt"
	"time"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/model"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/client"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/config"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/metadata"
)

type nextRecordProvider func() (ret int, ts interface{}, rec map[interface{}]interface{})

type Plugin struct {
	getConfigValue   func(string) string
	metadataProvider metadata.Provider

	keys *parseKeys

	client client.Client
}

func New(getConfigValue func(string) string, metadataProvider metadata.Provider, ingestionClient client.Client) (*Plugin, error) {
	p := &Plugin{
		getConfigValue:   getConfigValue,
		metadataProvider: metadataProvider,
	}

	keys := getParseKeys(getConfigValue, metadataProvider)
	p.keys = keys

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

func (p *Plugin) Transform(provider nextRecordProvider, tag string) map[model.Resource][]*model.Entry {
	resourceToEntries := make(map[model.Resource][]*model.Entry)

	for {
		ret, ts, record := provider()
		if ret != 0 {
			break
		}

		entry, res, err := p.entry(toTime(ts), record, tag)
		if err != nil {
			fmt.Printf("yc-logging: could not write entry %v because of error: %s\n", record, err.Error())
			continue
		}
		entries, ok := resourceToEntries[res]
		if ok {
			entries = append(entries, entry)
		} else {
			entries = []*model.Entry{entry}
		}
		resourceToEntries[res] = entries
	}

	return resourceToEntries
}

func (p *Plugin) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*model.Entry, model.Resource, error) {
	return p.keys.entry(ts, record, tag)
}
