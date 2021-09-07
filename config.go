package main

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	ingest "github.com/yandex-cloud/go-sdk/gen/logingestion"
)

func getConfigKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func getDestination(plugin unsafe.Pointer) (*logging.Destination, error) {
	const (
		keyFolderID = "folder_id"
		keyGroupID  = "group_id"
	)

	if groupID := getConfigKey(plugin, keyGroupID); len(groupID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigKey(plugin, keyFolderID); len(folderID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	return nil, fmt.Errorf("either %q or %q must be specified", keyGroupID, keyFolderID)
}

func getResource(plugin unsafe.Pointer) *logging.LogEntryResource {
	const (
		keyResourceType = "resource_type"
		keyResourceID   = "resource_id"
	)

	resourceType := getConfigKey(plugin, keyResourceType)
	resourceID := getConfigKey(plugin, keyResourceID)

	if len(resourceType)+len(resourceID) > 0 {
		return &logging.LogEntryResource{
			Type: resourceType,
			Id:   resourceID,
		}
	}
	return nil
}

func getDefaults(plugin unsafe.Pointer) (*logging.LogEntryDefaults, error) {
	const (
		keyDefaultLevel   = "default_level"
		keyDefaultPayload = "default_payload"
	)

	entryDefaults := new(logging.LogEntryDefaults)
	haveDefaults := false

	defaultLevel := getConfigKey(plugin, keyDefaultLevel)
	if len(defaultLevel) > 0 {
		level, err := levelFromString(defaultLevel)
		if err != nil {
			return nil, err
		}
		entryDefaults.Level = level
		haveDefaults = true
		println(fmt.Sprintf("yc-logging: will use %s as default level", level.String()))
	}

	defaultPayload := getConfigKey(plugin, keyDefaultPayload)
	if len(defaultPayload) > 0 {
		payload, err := payloadFromString(defaultPayload)
		if err != nil {
			return nil, err
		}
		entryDefaults.JsonPayload = payload
		haveDefaults = true
		data, _ := payload.MarshalJSON()
		println(fmt.Sprintf("yc-logging: will default payload:\n%s", string(data)))
	}

	if haveDefaults {
		return entryDefaults, nil
	}
	return nil, nil
}

func getParseKeys(plugin unsafe.Pointer) *parseKeys {
	const (
		keyLevelKey      = "level_key"
		keyMessageKey    = "message_key"
		keyMessageTagKey = "message_tag_key"
	)

	return &parseKeys{
		level:      getConfigKey(plugin, keyLevelKey),
		message:    getConfigKey(plugin, keyMessageKey),
		messageTag: getConfigKey(plugin, keyMessageTagKey),
	}
}

func getIngestionClient(plugin unsafe.Pointer) (*ingest.LogIngestionServiceClient, error) {
	const (
		keyAuthorization = "authorization"
		keyEndpoint      = "endpoint"
		defaultEndpoint  = "api.cloud.yandex.net:443"
	)

	authorization := getConfigKey(plugin, keyAuthorization)
	if authorization == "" {
		return nil, fmt.Errorf("authorization missing")
	}

	credentials, err := makeCredentials(authorization)
	if err != nil {
		return nil, err
	}

	endpoint := getConfigKey(plugin, keyEndpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: credentials,
		Endpoint:    endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating sdk: %s", err.Error())
	}
	return sdk.LogIngestion().LogIngestion(), nil
}
