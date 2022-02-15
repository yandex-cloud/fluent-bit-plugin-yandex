package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

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
