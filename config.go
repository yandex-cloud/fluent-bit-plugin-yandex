package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func getConfigKey(plugin unsafe.Pointer, key string) string {
	return strings.TrimSpace(output.FLBPluginConfigKey(plugin, key))
}

func getDestination(plugin unsafe.Pointer) (*logging.Destination, error) {
	const (
		keyFolderID = "folder_id"
		keyGroupID  = "group_id"
		urlFolderID = "http://" + ycsdk.InstanceMetadataAddr + "/computeMetadata/v1/yandex/folder-id"
	)

	if groupID := getConfigKey(plugin, keyGroupID); len(groupID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: groupID}}, nil
	}

	if folderID := getConfigKey(plugin, keyFolderID); len(folderID) > 0 {
		return &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: folderID}}, nil
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", urlFolderID, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make request to autodetect folder ID: %s", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get instance metadata to autodetect folder ID: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to autodetect folder ID returned status other than OK: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body returned by request to autodetect folder ID read failed: %s", err)
	}
	folderId := string(body)

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
