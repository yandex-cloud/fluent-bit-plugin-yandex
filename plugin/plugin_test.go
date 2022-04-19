package plugin

import (
	"testing"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/model"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/test"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

var configMap map[string]string
var getConfigValue = func(key string) string {
	val, ok := configMap[key]
	if ok {
		return val
	}
	return ""
}

func TestInit_AllConfig_GroupID_Success(t *testing.T) {
	configMap = map[string]string{
		"level_key":       "level",
		"message_key":     "message",
		"message_tag_key": "message_tag",
		"resource_type":   "resource_type",
		"resource_id":     "resource_id",
		"group_id":        "group_id",
		"default_level":   "INFO",
		"default_payload": "{}",
		"authorization":   "instance-service-account",
	}
	metadataProvider := test.MetadataProvider{}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, "level", plugin.keys.level)
	assert.Equal(t, "message", plugin.keys.message)
	assert.Equal(t, "message_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{"resource_type", [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{"resource_id", [][]string{}}, plugin.keys.resourceID)
}
func TestInit_AllConfigTemplated_GroupID_Success(t *testing.T) {
	configMap = map[string]string{
		"level_key":       "{{level}}",
		"message_key":     "{{message}}",
		"message_tag_key": "message_{{tag}}",
		"resource_type":   "resource_{{type}}",
		"resource_id":     "resource_{{id}}",
		"group_id":        "{{group_id}}",
		"default_level":   "{{default_level}}",
		"default_payload": "{{payload}}",
		"authorization":   "{{authorization}}",
	}
	metadataProvider := test.MetadataProvider{
		"level":         "metadata_level",
		"message":       "metadata_message",
		"tag":           "metadata_tag",
		"type":          "metadata_type",
		"id":            "metadata_id",
		"group_id":      "metadata_group_id",
		"default_level": "INFO",
		"payload":       "{}",
		"authorization": "instance-service-account",
	}
	client := &test.Client{}

	plugin, err := New(getConfigValue, metadataProvider, client)

	assert.Nil(t, err)
	assert.Equal(t, "metadata_level", plugin.keys.level)
	assert.Equal(t, "metadata_message", plugin.keys.message)
	assert.Equal(t, "message_metadata_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{"resource_metadata_type", [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{"resource_metadata_id", [][]string{}}, plugin.keys.resourceID)
}

func TestTransform_Success(t *testing.T) {
	records := []map[interface{}]interface{}{
		{"type": "1_type", "id": "1_id", "name": 10},
		{"type": "2_type", "id": "2_id", "name": 20},
	}
	var cur uint64 = 0
	var recordProvider = func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
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
		{"type": "1_type", "id": "1_id", "name": 10},
		{"type": "1_type", "id": "2_id", "name": 20},
		{"type": "1_type", "id": "2_id", "name": 21},
		{"type": "2_type", "id": "1_id", "name": 30},
		{"type": "2_type", "id": "1_id", "name": 31},
		{"type": "2_type", "id": "1_id", "name": 32},
		{"type": "2_type", "id": "2_id", "name": 40},
		{"type": "2_type", "id": "2_id", "name": 41},
		{"type": "2_type", "id": "2_id", "name": 42},
		{"type": "2_type", "id": "2_id", "name": 43},
	}
	var cur uint64 = 0
	var recordProvider = func() (ret int, ts interface{}, rec map[interface{}]interface{}) {
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
		},
	}

	resourceToEntries := plugin.Transform(recordProvider, "tag")

	expected := map[model.Resource][]*logging.IncomingLogEntry{
		{Type: "1_type", ID: "1_id"}: {{}},
		{Type: "1_type", ID: "2_id"}: {{}, {}},
		{Type: "2_type", ID: "1_id"}: {{}, {}, {}},
		{Type: "2_type", ID: "2_id"}: {{}, {}, {}, {}},
	}
	assert.NotNil(t, resourceToEntries)
	assert.Equal(t, len(expected), len(resourceToEntries))
	for k, v := range expected {
		actualV, ok := resourceToEntries[k]
		assert.True(t, ok)
		assert.Equal(t, len(v), len(actualV))
	}
}
