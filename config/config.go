package config

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/fluent/fluent-bit-go/output"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/metadata"
)

var (
	PluginVersion    string
	FluentBitVersion string
)

func GetKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func GetDestination(getConfigValue func(string) string, metadataProvider metadata.Provider) (*model.Destination, error) {
	const (
		keyFolderID         = "folder_id"
		keyGroupID          = "group_id"
		metadataKeyFolderID = "yandex/folder-id"
	)

	if groupID := getConfigValue(keyGroupID); len(groupID) > 0 {
		groupID = metadata.Parse(groupID, metadataProvider)
		return &model.Destination{LogGroupID: groupID}, nil
	}

	if folderID := getConfigValue(keyFolderID); len(folderID) > 0 {
		folderID = metadata.Parse(folderID, metadataProvider)
		return &model.Destination{FolderID: folderID}, nil
	}

	folderId, err := metadataProvider.GetValue(metadataKeyFolderID)
	if err != nil {
		return nil, err
	}

	return &model.Destination{FolderID: folderId}, nil
}

func GetDefaults(getConfigValue func(string) string, metadataProvider metadata.Provider) (*model.Defaults, error) {
	const (
		keyDefaultLevel   = "default_level"
		keyDefaultPayload = "default_payload"
	)

	entryDefaults := new(model.Defaults)
	haveDefaults := false

	defaultLevel := getConfigValue(keyDefaultLevel)
	if len(defaultLevel) > 0 {
		defaultLevel = metadata.Parse(defaultLevel, metadataProvider)
		entryDefaults.Level = defaultLevel
		haveDefaults = true
	}

	defaultPayload := getConfigValue(keyDefaultPayload)
	if len(defaultPayload) > 0 {
		var err error
		defaultPayload = metadata.Parse(defaultPayload, metadataProvider)
		payload, err := payloadFromString(defaultPayload)
		if err != nil {
			return nil, err
		}
		entryDefaults.JSONPayload = payload
		haveDefaults = true
	}

	if haveDefaults {
		return entryDefaults, nil
	}
	return nil, nil
}

func GetAuthorization(getConfigValue func(string) string, metadataProvider metadata.Provider) (string, error) {
	const keyAuthorization = "authorization"

	authorization := getConfigValue(keyAuthorization)
	if authorization == "" {
		return "", fmt.Errorf("authorization missing")
	}

	return metadata.Parse(authorization, metadataProvider), nil
}

func GetEndpoint(getConfigValue func(string) string) string {
	const (
		keyEndpoint     = "endpoint"
		defaultEndpoint = "api.cloud.yandex.net:443"
	)

	endpoint := getConfigValue(keyEndpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	return endpoint
}

func GetCAFileName(getConfigValue func(string) string) string {
	const CAFileNameKey = "ca_file"

	return getConfigValue(CAFileNameKey)
}

func payloadFromString(payload string) (*structpb.Struct, error) {
	result := new(structpb.Struct)
	err := result.UnmarshalJSON([]byte(payload))
	if err != nil {
		return nil, err
	}
	return result, nil
}
