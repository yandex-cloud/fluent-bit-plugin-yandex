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

func getMetadataValue(key string) (string, error) {
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
