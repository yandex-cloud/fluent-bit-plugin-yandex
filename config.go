package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func getConfigKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func getMetadata(key string) (string, error) {
	const (
		keyMetadataUrlEnv = "YC_METADATA_URL"
		urlSuffix         = "/computeMetadata/v1/"
		requestTimeout    = 5 * time.Second
	)

	metadataEndpoint := os.Getenv(keyMetadataUrlEnv)
	if len(metadataEndpoint) == 0 {
		metadataEndpoint = "http://" + ycsdk.InstanceMetadataAddr
	}
	urlMetadata := metadataEndpoint + urlSuffix + key

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlMetadata, nil)
	if err != nil {
		return "", fmt.Errorf("could not make request to autodetect folder ID: %s", err.Error())
	}
	req.Header.Set("Metadata-Flavor", "Google")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("could not get instance metadata to autodetect folder ID: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to autodetect folder ID returned status other than OK: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response body returned by request to autodetect folder ID read failed: %s", err.Error())
	}

	return string(body), nil
}

func getDestination(plugin unsafe.Pointer) (*logging.Destination, error) {
	const (
		keyFolderID         = "folder_id"
		keyGroupID          = "group_id"
		metadataKeyFolderID = "yandex/folder-id"
	)

	if groupID := getConfigKey(plugin, keyGroupID); len(groupID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigKey(plugin, keyFolderID); len(folderID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	folderId, err := getMetadata(metadataKeyFolderID)
	if err != nil {
		return nil, err
	}

	return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderId}}, nil
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
		fmt.Printf("yc-logging: will use %s as default level\n", level.String())
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
		fmt.Printf("yc-logging: will default payload:\n%s\n", string(data))
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
