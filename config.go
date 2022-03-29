package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func getConfigKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func getDestination(getConfigValue func(string) string, metadataProvider MetadataProvider) (*logging.Destination, error) {
	const (
		keyFolderID         = "folder_id"
		keyGroupID          = "group_id"
		metadataKeyFolderID = "yandex/folder-id"
	)

	if groupID := getConfigValue(keyGroupID); len(groupID) > 0 {
		groupID, err := parseWithMetadata(groupID, metadataProvider)
		if err != nil {
			return nil, err
		}
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigValue(keyFolderID); len(folderID) > 0 {
		folderID, err := parseWithMetadata(folderID, metadataProvider)
		if err != nil {
			return nil, err
		}
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	folderId, err := metadataProvider.getValue(metadataKeyFolderID)
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
		defaultLevel, err = parseWithMetadata(defaultLevel, metadataProvider)
		if err != nil {
			return nil, err
		}
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
		defaultPayload, err = parseWithMetadata(defaultPayload, metadataProvider)
		if err != nil {
			return nil, err
		}
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

	level, err := parseWithMetadata(getConfigValue(keyLevelKey), metadataProvider)
	if err != nil {
		return nil, err
	}
	message, err := parseWithMetadata(getConfigValue(keyMessageKey), metadataProvider)
	if err != nil {
		return nil, err
	}
	messageTag, err := parseWithMetadata(getConfigValue(keyMessageTagKey), metadataProvider)
	if err != nil {
		return nil, err
	}

	resourceType, err := parseWithMetadata(getConfigValue(keyResourceType), metadataProvider)
	if err != nil {
		return nil, err
	}
	resourceID, err := parseWithMetadata(getConfigValue(keyResourceID), metadataProvider)
	if err != nil {
		return nil, err
	}

	return &parseKeys{
		level:        level,
		message:      message,
		messageTag:   messageTag,
		resourceType: newTemplate(resourceType),
		resourceID:   newTemplate(resourceID),
	}, nil
}
