package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/plugin"
)

var configMap map[string]string
var getConfigValue = func(key string) string {
	val, ok := configMap[key]
	if ok {
		return val
	}
	return ""
}

func TestPlugin_Success(t *testing.T) {
	configMap = map[string]string{
		"level_key":       "{{level}}",
		"message_key":     "{{message}}",
		"message_tag_key": "{{tag}}",
		"resource_type":   "{type}_{{resource_type}}",
		"resource_id":     "{id}_{{resource_id}}",
		"group_id":        "{{group_id}}",
		"default_level":   "{{default_level}}",
		"default_payload": "{{payload}}",
		"authorization":   "{{authorization}}",
	}
	metadataProvider := plugin.TestMetadataProvider{
		"level":         "metadata_level",
		"message":       "metadata_message",
		"tag":           "metadata_tag",
		"resource_type": "type",
		"resource_id":   "id",
		"group_id":      "metadata_group_id",
		"default_level": "INFO",
		"payload":       "{}",
		"authorization": "instance-service-account",
	}

	impl, err := plugin.New(getConfigValue, metadataProvider)

	assert.Nil(t, err)

	records := []map[interface{}]interface{}{
		{"type": "1", "id": "1", "name": 10},
		{"type": "2", "id": "2", "name": 20},
	}
	var cur uint64 = 0
	var recordProvider = func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
		if int(cur) >= len(records) {
			return 1, nil, nil
		}
		cur++
		return 0, cur - 1, records[cur-1]
	}

	resourceToEntries := impl.Transform(recordProvider, "tag")

	assert.NotNil(t, resourceToEntries)
	assert.Equal(t, 2, len(resourceToEntries))
	types := make([]string, 0)
	for res := range resourceToEntries {
		types = append(types, res.LogEntryResource().Type)
	}
	sort.Strings(types)
	assert.Equal(t, []string{"1_type", "2_type"}, types)
	for resource, entries := range resourceToEntries {
		resource := resource.LogEntryResource()
		switch resource.Type {
		case "1_type":
			assert.Equal(t, "1_id", resource.Id)
			assert.Equal(t, 1, len(entries))
			actualPayload := entries[0].JsonPayload.AsMap()
			assert.Equal(t, float64(10), actualPayload["name"])
			assert.Equal(t, "1", actualPayload["type"])
			assert.Equal(t, "1", actualPayload["id"])
		case "2_type":
			assert.Equal(t, "2_id", resource.Id)
			assert.Equal(t, 1, len(entries))
			actualPayload := entries[0].JsonPayload.AsMap()
			assert.Equal(t, float64(20), actualPayload["name"])
			assert.Equal(t, "2", actualPayload["type"])
			assert.Equal(t, "2", actualPayload["id"])
		}
	}
}
