package plugin

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func getDestination(getConfigValue func(string) string, metadataProvider MetadataProvider) (*logging.Destination, error) {
	const (
		keyFolderID         = "folder_id"
		keyGroupID          = "group_id"
		metadataKeyFolderID = "yandex/folder-id"
	)

	if groupID := getConfigValue(keyGroupID); len(groupID) > 0 {
		groupID = parseWithMetadata(groupID, metadataProvider)
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigValue(keyFolderID); len(folderID) > 0 {
		folderID = parseWithMetadata(folderID, metadataProvider)
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	folderId, err := metadataProvider.GetValue(metadataKeyFolderID)
	if err != nil {
		return nil, err
	}

	return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderId}}, nil
}

func getDefaults(getConfigValue func(string) string, metadataProvider MetadataProvider) (*logging.LogEntryDefaults, error) {
	const (
		keyDefaultLevel   = "default_level"
		keyDefaultPayload = "default_payload"
	)

	entryDefaults := new(logging.LogEntryDefaults)
	haveDefaults := false

	defaultLevel := getConfigValue(keyDefaultLevel)
	if len(defaultLevel) > 0 {
		var err error
		defaultLevel = parseWithMetadata(defaultLevel, metadataProvider)
		level, err := levelFromString(defaultLevel)
		if err != nil {
			return nil, err
		}
		entryDefaults.Level = level
		haveDefaults = true
		fmt.Printf("yc-logging: will use %s as default level\n", level.String())
	}

	defaultPayload := getConfigValue(keyDefaultPayload)
	if len(defaultPayload) > 0 {
		var err error
		defaultPayload = parseWithMetadata(defaultPayload, metadataProvider)
		payload, err := payloadFromString(defaultPayload)
		if err != nil {
			return nil, err
		}
		entryDefaults.JsonPayload = payload
		haveDefaults = true
		data, _ := payload.MarshalJSON()
		fmt.Printf("yc-logging: will default payload:\n%s\n", string(data))
	}

	if haveDefaults {
		return entryDefaults, nil
	}
	return nil, nil
}

func getParseKeys(getConfigValue func(string) string, metadataProvider MetadataProvider) (*parseKeys, error) {
	const (
		keyLevelKey      = "level_key"
		keyMessageKey    = "message_key"
		keyMessageTagKey = "message_tag_key"
		keyResourceType  = "resource_type"
		keyResourceID    = "resource_id"
	)

	level := parseWithMetadata(getConfigValue(keyLevelKey), metadataProvider)
	message := parseWithMetadata(getConfigValue(keyMessageKey), metadataProvider)
	messageTag := parseWithMetadata(getConfigValue(keyMessageTagKey), metadataProvider)

	resourceType := parseWithMetadata(getConfigValue(keyResourceType), metadataProvider)
	resourceID := parseWithMetadata(getConfigValue(keyResourceID), metadataProvider)

	return &parseKeys{
		level:        level,
		message:      message,
		messageTag:   messageTag,
		resourceType: newTemplate(resourceType),
		resourceID:   newTemplate(resourceID),
	}, nil
}

func getAuthorization(getConfigValue func(string) string, metadataProvider MetadataProvider) (string, error) {
	const keyAuthorization = "authorization"

	authorization := getConfigValue(keyAuthorization)
	if authorization == "" {
		return "", fmt.Errorf("authorization missing")
	}

	return parseWithMetadata(authorization, metadataProvider), nil
}

func getEndpoint(getConfigValue func(string) string) string {
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

func getCAFileName(getConfigValue func(string) string) string {
	const CAFileNameKey = "ca_file"

	return getConfigValue(CAFileNameKey)
}
