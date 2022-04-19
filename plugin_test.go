package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/plugin"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/test"
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
	}
	metadataProvider := test.MetadataProvider{
		"level":         "metadata_level",
		"message":       "metadata_message",
		"tag":           "metadata_tag",
		"resource_type": "type",
		"resource_id":   "id",
	}
	client := &test.Client{}

	impl, err := plugin.New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)

	records := []map[interface{}]interface{}{
		{"type": "1", "id": "1", "name": 10, "metadata_message": "message_10", "metadata_level": "ERROR"},
		{"type": "2", "id": "2", "name": 20, "metadata_message": "message_20", "metadata_level": "WARN"},
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
		types = append(types, res.Type)
	}
	sort.Strings(types)
	assert.Equal(t, []string{"1_type", "2_type"}, types)
	for resource, entries := range resourceToEntries {
		resource := resource
		switch resource.Type {
		case "1_type":
			assert.Equal(t, "1_id", resource.ID)
			assert.Equal(t, 1, len(entries))
			assert.Equal(t, int64(0), entries[0].Timestamp.Unix())
			assert.Equal(t, "ERROR", entries[0].Level)
			assert.Equal(t, "message_10", entries[0].Message)
			actualPayload := entries[0].JSONPayload.AsMap()
			assert.Equal(t, float64(10), actualPayload["name"])
			assert.Equal(t, "tag", actualPayload["metadata_tag"])
			assert.Equal(t, "1", actualPayload["type"])
			assert.Equal(t, "1", actualPayload["id"])
		case "2_type":
			assert.Equal(t, "2_id", resource.ID)
			assert.Equal(t, 1, len(entries))
			assert.Equal(t, int64(1), entries[0].Timestamp.Unix())
			assert.Equal(t, "WARN", entries[0].Level)
			assert.Equal(t, "message_20", entries[0].Message)
			actualPayload := entries[0].JSONPayload.AsMap()
			assert.Equal(t, float64(20), actualPayload["name"])
			assert.Equal(t, "tag", actualPayload["metadata_tag"])
			assert.Equal(t, "2", actualPayload["type"])
			assert.Equal(t, "2", actualPayload["id"])
		}
	}
}
