package plugin

import (
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/metadata"
	"strings"
)

func getParseKeys(getConfigValue func(string) string, metadataProvider metadata.Provider) *parseKeys {
	const (
		keyLevelKey        = "level_key"
		keyMessageKey      = "message_key"
		keyMessageKeysList = "message_keys"
		keyMessageTagKey   = "message_tag_key"
		keyResourceType    = "resource_type"
		keyResourceID      = "resource_id"
		keySteamName       = "stream_name"
	)

	messageKeysListString := metadata.Parse(getConfigValue(keyMessageKeysList), metadataProvider)
	if messageKeysListString == "" {
		messageKeysListString = metadata.Parse(getConfigValue(keyMessageKey), metadataProvider)
	}
	messageKeysList := strings.Split(messageKeysListString, " ")
	messageKeysListMap := make(map[string]struct{}, len(messageKeysList))
	for _, key := range messageKeysList {
		messageKeysListMap[key] = struct{}{}
	}

	level := metadata.Parse(getConfigValue(keyLevelKey), metadataProvider)
	messageTag := metadata.Parse(getConfigValue(keyMessageTagKey), metadataProvider)

	resourceType := metadata.Parse(getConfigValue(keyResourceType), metadataProvider)
	resourceID := metadata.Parse(getConfigValue(keyResourceID), metadataProvider)
	streamName := metadata.Parse(getConfigValue(keySteamName), metadataProvider)

	return &parseKeys{
		level:        level,
		messageKeys:  messageKeysListMap,
		messageTag:   messageTag,
		resourceType: newTemplate(resourceType),
		resourceID:   newTemplate(resourceID),
		streamName:   newTemplate(streamName),
	}
}
