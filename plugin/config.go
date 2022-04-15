package plugin

import (
	"fmt"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/metadata"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/util"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func getDestination(getConfigValue func(string) string, metadataProvider metadata.MetadataProvider) (*logging.Destination, error) {
	const (
		keyFolderID         = "folder_id"
		keyGroupID          = "group_id"
		metadataKeyFolderID = "yandex/folder-id"
	)

	if groupID := getConfigValue(keyGroupID); len(groupID) > 0 {
		groupID = metadata.Parse(groupID, metadataProvider)
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigValue(keyFolderID); len(folderID) > 0 {
		folderID = metadata.Parse(folderID, metadataProvider)
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	folderId, err := metadataProvider.GetValue(metadataKeyFolderID)
	if err != nil {
		return nil, err
	}

	return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderId}}, nil
}

func getDefaults(getConfigValue func(string) string, metadataProvider metadata.MetadataProvider) (*logging.LogEntryDefaults, error) {
	const (
		keyDefaultLevel   = "default_level"
		keyDefaultPayload = "default_payload"
	)

	entryDefaults := new(logging.LogEntryDefaults)
	haveDefaults := false

	defaultLevel := getConfigValue(keyDefaultLevel)
	if len(defaultLevel) > 0 {
		var err error
		defaultLevel = metadata.Parse(defaultLevel, metadataProvider)
		level, err := util.LevelFromString(defaultLevel)
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
		defaultPayload = metadata.Parse(defaultPayload, metadataProvider)
		payload, err := util.PayloadFromString(defaultPayload)
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

func getParseKeys(getConfigValue func(string) string, metadataProvider metadata.MetadataProvider) *parseKeys {
	const (
		keyLevelKey      = "level_key"
		keyMessageKey    = "message_key"
		keyMessageTagKey = "message_tag_key"
		keyResourceType  = "resource_type"
		keyResourceID    = "resource_id"
	)

	level := metadata.Parse(getConfigValue(keyLevelKey), metadataProvider)
	message := metadata.Parse(getConfigValue(keyMessageKey), metadataProvider)
	messageTag := metadata.Parse(getConfigValue(keyMessageTagKey), metadataProvider)

	resourceType := metadata.Parse(getConfigValue(keyResourceType), metadataProvider)
	resourceID := metadata.Parse(getConfigValue(keyResourceID), metadataProvider)

	return &parseKeys{
		level:        level,
		message:      message,
		messageTag:   messageTag,
		resourceType: util.NewTemplate(resourceType),
		resourceID:   util.NewTemplate(resourceID),
	}
}
