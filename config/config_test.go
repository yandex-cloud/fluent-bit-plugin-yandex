package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/model"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/test"
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

func TestGetDestination_GroupID_Success(t *testing.T) {
	configMap = map[string]string{
		"group_id": "abcdef",
	}
	metadataProvider := test.MetadataProvider{}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{LogGroupID: "abcdef"}, destination)
}
func TestGetDestination_GroupIDTemplated_Success(t *testing.T) {
	configMap = map[string]string{
		"group_id": "{{group}}",
	}
	metadataProvider := test.MetadataProvider{
		"group": "abcdef",
	}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{LogGroupID: "abcdef"}, destination)
}
func TestGetDestination_GroupIDFolderID_Success(t *testing.T) {
	configMap = map[string]string{
		"group_id":  "abcdef",
		"folder_id": "qwerty",
	}
	metadataProvider := test.MetadataProvider{}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{LogGroupID: "abcdef"}, destination)
}
func TestGetDestination_FolderID_Success(t *testing.T) {
	configMap = map[string]string{
		"folder_id": "qwerty",
	}
	metadataProvider := test.MetadataProvider{}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{FolderID: "qwerty"}, destination)
}
func TestGetDestination_FolderIDTemplated_Success(t *testing.T) {
	configMap = map[string]string{
		"folder_id": "{{folder}}",
	}
	metadataProvider := test.MetadataProvider{
		"folder": "qwerty",
	}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{FolderID: "qwerty"}, destination)
}
func TestGetDestination_FolderIDAutoDetection_Success(t *testing.T) {
	configMap = map[string]string{}
	metadataProvider := test.MetadataProvider{
		"yandex/folder-id": "folder-id",
	}

	destination, err := GetDestination(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Destination{FolderID: "folder-id"}, destination)
}
func TestGetDestination_FolderIDAutoDetection_Fail(t *testing.T) {
	configMap = map[string]string{}
	metadataProvider := test.MetadataProvider{}

	_, err := GetDestination(getConfigValue, metadataProvider)

	assert.NotNil(t, err)
}

func TestGetDefaults_Empty_Success(t *testing.T) {
	configMap = map[string]string{}
	metadataProvider := test.MetadataProvider{}

	defaults, err := GetDefaults(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Nil(t, defaults)
}
func TestGetDefaults_LevelOnly_Success(t *testing.T) {
	configMap = map[string]string{
		"default_level": "INFO",
	}
	metadataProvider := test.MetadataProvider{}

	defaults, err := GetDefaults(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, &model.Defaults{Level: "INFO"}, defaults)
}
func TestGetDefaults_PayloadOnly_Success(t *testing.T) {
	configMap = map[string]string{
		"default_payload": "{}",
	}
	metadataProvider := test.MetadataProvider{}

	defaults, err := GetDefaults(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, "", defaults.Level)
	assert.Equal(t, map[string]*structpb.Value{}, defaults.JSONPayload.Fields)
}
func TestGetDefaults_PayloadOnly_Fail(t *testing.T) {
	configMap = map[string]string{
		"default_payload": "{incorrect json\"",
	}
	metadataProvider := test.MetadataProvider{}

	_, err := GetDefaults(getConfigValue, metadataProvider)

	assert.NotNil(t, err)
}
func TestGetDefaults_Success(t *testing.T) {
	configMap = map[string]string{
		"default_level":   "INFO",
		"default_payload": "{}",
	}
	metadataProvider := test.MetadataProvider{}

	defaults, err := GetDefaults(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, "INFO", defaults.Level)
	assert.Equal(t, map[string]*structpb.Value{}, defaults.JSONPayload.Fields)
}
func TestGetDefaults_Templated_Success(t *testing.T) {
	configMap = map[string]string{
		"default_level":   "{{level}}",
		"default_payload": "{{payload}}",
	}
	metadataProvider := test.MetadataProvider{
		"level":   "INFO",
		"payload": "{}",
	}

	defaults, err := GetDefaults(getConfigValue, metadataProvider)

	assert.Nil(t, err)
	assert.Equal(t, "INFO", defaults.Level)
	assert.Equal(t, map[string]*structpb.Value{}, defaults.JSONPayload.Fields)
}
