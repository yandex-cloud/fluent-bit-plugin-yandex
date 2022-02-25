package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/startdusk/strnaming"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func getMetadataUrl() string {
	const (
		keyMetadataUrlEnv = "YC_METADATA_URL"
		urlSuffix         = "/computeMetadata/v1/"
	)
	metadataEndpoint := os.Getenv(keyMetadataUrlEnv)
	if len(metadataEndpoint) == 0 {
		metadataEndpoint = "http://" + ycsdk.InstanceMetadataAddr
	}
	return metadataEndpoint + urlSuffix
}

func getMetadataValue(key string) (string, error) {
	const requestTimeout = 5 * time.Second

	urlMetadata := getMetadataUrl() + key

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlMetadata, nil)
	if err != nil {
		return "", fmt.Errorf("could not make request to get metadata value by key %q: %s", key, err.Error())
	}
	req.Header.Set("Metadata-Flavor", "Google")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("could not get instance metadata value by key %q: %s", key, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to get metadata value by key %q returned status other than OK: %s", key, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response body returned by request to get metadata value by key %q read failed: %s", key, err.Error())
	}

	return string(body), nil
}

func getAllMetadata() (*structpb.Struct, error) {
	const (
		queryParam     = "?recursive=true"
		requestTimeout = 5 * time.Second
	)

	urlMetadata := getMetadataUrl() + queryParam

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlMetadata, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make request to get all metadata: %s", err.Error())
	}
	req.Header.Set("Metadata-Flavor", "Google")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not get all metadata: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to get all metadata returned status other than OK: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body returned by request to get all metadata read failed: %s", err.Error())
	}

	metadataStruct := new(structpb.Struct)
	err = metadataStruct.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body returned by request to get all metadata: %s", err.Error())
	}

	return metadataStruct, nil
}

func getCachedMetadataValue(metadata *structpb.Struct, key string) (string, error) {
	toCamel := strnaming.NewCamel()
	toCamel.WithDelimiter('-')

	key = toCamel.Convert(key)
	path := strings.Split(key, "/")

	value, err := getValue(metadata, path)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata value by key %q because of error: %s", key, err.Error())
	}

	if _, ok := value.GetKind().(*structpb.Value_StringValue); ok {
		return value.GetStringValue(), nil
	}

	content, err := value.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON from value by key %q: %s", key, err.Error())
	}
	return string(content), nil
}
