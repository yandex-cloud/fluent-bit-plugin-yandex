package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/protobuf/types/known/structpb"
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
	metadataProvider := TestMetadataProvider{}

	plugin, err := New(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: "group_id"}}, plugin.destination)
	assert.Equal(t, logging.LogLevel_INFO, plugin.defaults.Level)
	assert.Equal(t, map[string]*structpb.Value{}, plugin.defaults.JsonPayload.Fields)
	assert.Equal(t, "level", plugin.keys.level)
	assert.Equal(t, "message", plugin.keys.message)
	assert.Equal(t, "message_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{format: "resource_type", keys: [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{format: "resource_id", keys: [][]string{}}, plugin.keys.resourceID)
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
		"authorization":   "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{
		"level":         "metadata_level",
		"message":       "metadata_message",
		"tag":           "metadata_tag",
		"type":          "metadata_type",
		"id":            "metadata_id",
		"group_id":      "metadata_group_id",
		"default_level": "INFO",
		"payload":       "{}",
	}

	plugin, err := New(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: "metadata_group_id"}}, plugin.destination)
	assert.Equal(t, logging.LogLevel_INFO, plugin.defaults.Level)
	assert.Equal(t, map[string]*structpb.Value{}, plugin.defaults.JsonPayload.Fields)
	assert.Equal(t, "metadata_level", plugin.keys.level)
	assert.Equal(t, "metadata_message", plugin.keys.message)
	assert.Equal(t, "message_metadata_tag", plugin.keys.messageTag)
	assert.Equal(t, &template{format: "resource_metadata_type", keys: [][]string{}}, plugin.keys.resourceType)
	assert.Equal(t, &template{format: "resource_metadata_id", keys: [][]string{}}, plugin.keys.resourceID)
}
func TestInit_FolderIDTemplated_Success(t *testing.T) {
	configMap = map[string]string{
		"folder_id":     "{{folder_id}}",
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{
		"folder_id": "folder_id",
	}

	plugin, err := New(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: "folder_id"}}, plugin.destination)
}
func TestInit_FolderIDAutodetection_Success(t *testing.T) {
	configMap = map[string]string{
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{
		"yandex/folder-id": "folder-id",
	}

	plugin, err := New(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: "folder-id"}}, plugin.destination)
}
func TestInit_FolderIDAutodetection_Fail(t *testing.T) {
	configMap = map[string]string{
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{}

	_, err := New(getConfigValue, metadataProvider)

	assert.NotNil(t, err)
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
	actual1 := resourceToEntries[Resource{resourceType: "1_type", resourceID: "1_id"}]
	assert.Equal(t, 1, len(actual1))
	actualPayload1 := actual1[0].JsonPayload.AsMap()
	assert.Equal(t, float64(10), actualPayload1["name"])
	assert.Equal(t, "1_type", actualPayload1["type"])
	assert.Equal(t, "1_id", actualPayload1["id"])
	actual2 := resourceToEntries[Resource{resourceType: "2_type", resourceID: "2_id"}]
	assert.Equal(t, 1, len(actual2))
	actualPayload2 := actual2[0].JsonPayload.AsMap()
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

	expected := map[Resource][]*logging.IncomingLogEntry{
		{resourceType: "1_type", resourceID: "1_id"}: {{}},
		{resourceType: "1_type", resourceID: "2_id"}: {{}, {}},
		{resourceType: "2_type", resourceID: "1_id"}: {{}, {}, {}},
		{resourceType: "2_type", resourceID: "2_id"}: {{}, {}, {}, {}},
	}
	assert.NotNil(t, resourceToEntries)
	assert.Equal(t, len(expected), len(resourceToEntries))
	for k, v := range expected {
		actualV, ok := resourceToEntries[k]
		assert.True(t, ok)
		assert.Equal(t, len(v), len(actualV))
	}
}
