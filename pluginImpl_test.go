package main

import (
	"testing"

	"github.com/fluent/fluent-bit-go/output"

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

	plugin := new(pluginImpl)
	_, err := plugin.init(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: "group_id"}}, plugin.destination, "incorrect destination")
	assert.Equal(t, logging.LogLevel_INFO, plugin.defaults.Level, "incorrect default level")
	assert.Equal(t, map[string]*structpb.Value{}, plugin.defaults.JsonPayload.Fields, "incorrect default payload")
	assert.Equal(t, "level", plugin.keys.level, "incorrect level key")
	assert.Equal(t, "message", plugin.keys.message, "incorrect message key")
	assert.Equal(t, "message_tag", plugin.keys.messageTag, "incorrect message tag key")
	assert.Equal(t, &template{format: "resource_type", keys: [][]string{}}, plugin.keys.resourceType, "incorrect resource type")
	assert.Equal(t, &template{format: "resource_id", keys: [][]string{}}, plugin.keys.resourceID, "incorrect resource id")
}

func TestInit_FolderID_Success(t *testing.T) {
	configMap = map[string]string{
		"folder_id":     "folder_id",
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{}

	plugin := new(pluginImpl)
	_, err := plugin.init(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: "folder_id"}}, plugin.destination, "incorrect destination")
}

func TestInit_FolderIDAutodetection_Success(t *testing.T) {
	configMap = map[string]string{
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{
		"yandex/folder-id": "folder-id",
	}

	plugin := new(pluginImpl)
	_, err := plugin.init(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: "folder-id"}}, plugin.destination, "incorrect destination")
}

func TestInit_FolderIDAutodetection_Fail(t *testing.T) {
	configMap = map[string]string{
		"authorization": "instance-service-account",
	}
	metadataProvider := TestMetadataProvider{}

	plugin := new(pluginImpl)
	code, err := plugin.init(getConfigValue, metadataProvider)

	assert.Equal(t, output.FLB_ERROR, code)
	assert.NotNil(t, err)
}
