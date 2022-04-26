package plugin

import (
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/metadata"
)

func getParseKeys(getConfigValue func(string) string, metadataProvider metadata.Provider) *parseKeys {
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
		resourceType: newTemplate(resourceType),
		resourceID:   newTemplate(resourceID),
	}
}
