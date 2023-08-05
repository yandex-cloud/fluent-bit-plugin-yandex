package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/test"
)

var (
	configMap      map[string]string
	getConfigValue = func(key string) string {
		val, ok := configMap[key]
		if ok {
			return val
		}
		return ""
	}
)

func TestInit_AllConfig_Success(t *testing.T) {
	configMap = map[string]string{
		"level_key":       "level",
		"message_key":     "message",
		"message_tag_key": "message_tag",
		"resource_type":   "resource_type",
		"resource_id":     "resource_id",
		"stream_name":     "stream_name",
	}
	metadataProvider := test.MetadataProvider{}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, "level", plugin.keys.level)
	assert.Equal(t, map[string]struct{}{
		"message": {},
	}, plugin.keys.messageKeys)
	assert.Equal(t, "message_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{"resource_type", [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{"resource_id", [][]string{}}, plugin.keys.resourceID)
	assert.Equal(t, &template{"stream_name", [][]string{}}, plugin.keys.streamName)
}

func TestInit_AllConfigMultipleMessageKeys_Success(t *testing.T) {
	configMap = map[string]string{
		"message_keys": "msg message log",
	}
	metadataProvider := test.MetadataProvider{}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{
		"msg":     {},
		"message": {},
		"log":     {},
	}, plugin.keys.messageKeys)
}

func TestInit_AllConfigTemplated_Success(t *testing.T) {
	configMap = map[string]string{
		"level_key":       "{{level}}",
		"message_key":     "{{message}}",
		"message_tag_key": "message_{{tag}}",
		"resource_type":   "resource_{{type}}",
		"resource_id":     "resource_{{id}}",
	}
	metadataProvider := test.MetadataProvider{
		"level":   "metadata_level",
		"message": "metadata_message",
		"tag":     "metadata_tag",
		"type":    "metadata_type",
		"id":      "metadata_id",
	}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, "metadata_level", plugin.keys.level)
	assert.Equal(t, map[string]struct{}{
		"metadata_message": {},
	}, plugin.keys.messageKeys)
	assert.Equal(t, "message_metadata_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{"resource_metadata_type", [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{"resource_metadata_id", [][]string{}}, plugin.keys.resourceID)
}

func TestInit_AllConfigTemplatedMultipleMessageKeys_Success(t *testing.T) {
	configMap = map[string]string{
		"message_keys": "{{message}}",
	}
	metadataProvider := test.MetadataProvider{
		"message": "metadata_msg metadata_message metadata_log",
	}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{
		"metadata_msg":     {},
		"metadata_message": {},
		"metadata_log":     {},
	}, plugin.keys.messageKeys)
}

func TestTransform_Success(t *testing.T) {
	records := []map[interface{}]interface{}{
		{"type": "1_type", "id": "1_id", "name": 10, "stream": "stream1"},
		{"type": "2_type", "id": "2_id", "name": 20, "stream": "stream1"},
	}
	var cur uint64
	recordProvider := func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
		if int(cur) >= len(records) {
			return 1, nil, nil
		}
		cur++
		return 0, cur - 1, records[cur-1]
	}
	plugin := Plugin{
		keys: &parseKeys{
			resourceType: newTemplate("{type}"),
			resourceID:   newTemplate("{id}"),
			streamName:   newTemplate("{stream}"),
		},
	}

	resourceToEntries := plugin.Transform(recordProvider, "tag")

	assert.NotNil(t, resourceToEntries)
	actual1 := resourceToEntries[model.Resource{Type: "1_type", ID: "1_id"}]
	assert.Equal(t, 1, len(actual1))
	actualPayload1 := actual1[0].JSONPayload.AsMap()
	assert.Equal(t, float64(10), actualPayload1["name"])
	assert.Equal(t, "1_type", actualPayload1["type"])
	assert.Equal(t, "1_id", actualPayload1["id"])
	actual2 := resourceToEntries[model.Resource{Type: "2_type", ID: "2_id"}]
	assert.Equal(t, 1, len(actual2))
	actualPayload2 := actual2[0].JSONPayload.AsMap()
	assert.Equal(t, float64(20), actualPayload2["name"])
	assert.Equal(t, "2_type", actualPayload2["type"])
	assert.Equal(t, "2_id", actualPayload2["id"])
}

func TestTransform_IdentifyingResource_Success(t *testing.T) {
	records := []map[interface{}]interface{}{
		{"type": "1_type", "id": "1_id", "name": 10, "stream": "stream1"},
		{"type": "1_type", "id": "2_id", "name": 20, "stream": "stream1"},
		{"type": "1_type", "id": "2_id", "name": 21, "stream": "stream1"},
		{"type": "2_type", "id": "1_id", "name": 30, "stream": "stream1"},
		{"type": "2_type", "id": "1_id", "name": 31, "stream": "stream1"},
		{"type": "2_type", "id": "1_id", "name": 32, "stream": "stream1"},
		{"type": "2_type", "id": "2_id", "name": 40, "stream": "stream1"},
		{"type": "2_type", "id": "2_id", "name": 41, "stream": "stream1"},
		{"type": "2_type", "id": "2_id", "name": 42, "stream": "stream1"},
		{"type": "2_type", "id": "2_id", "name": 43, "stream": "stream1"},
	}
	var cur uint64
	recordProvider := func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
		if int(cur) >= len(records) {
			return 1, nil, nil
		}
		cur++
		return 0, cur - 1, records[cur-1]
	}
	plugin := Plugin{
		keys: &parseKeys{
			resourceType: newTemplate("{type}"),
			resourceID:   newTemplate("{id}"),
			streamName:   newTemplate("{stream}"),
		},
	}

	resourceToEntries := plugin.Transform(recordProvider, "tag")

	expected := map[model.Resource][]*logging.IncomingLogEntry{
		{Type: "1_type", ID: "1_id"}: {{StreamName: "stream1"}},
		{Type: "1_type", ID: "2_id"}: {{StreamName: "stream1"}, {StreamName: "stream1"}},
		{Type: "2_type", ID: "1_id"}: {{StreamName: "stream1"}, {StreamName: "stream1"}, {StreamName: "stream1"}},
		{Type: "2_type", ID: "2_id"}: {{StreamName: "stream1"}, {StreamName: "stream1"}, {StreamName: "stream1"}, {StreamName: "stream1"}},
	}
	assert.NotNil(t, resourceToEntries)
	assert.Equal(t, len(expected), len(resourceToEntries))
	for k, v := range expected {
		actualV, ok := resourceToEntries[k]
		assert.True(t, ok)
		assert.Equal(t, len(v), len(actualV))
	}
}
