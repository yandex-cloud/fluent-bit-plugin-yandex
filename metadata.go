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

type MetadataProvider interface {
	getMetadataValue(key string) (string, error)
}

type CachingMetadataProvider struct {
	cache *structpb.Struct
}

func (mp *CachingMetadataProvider) getMetadataValue(key string) (string, error) {
	if mp.cache == nil {
		err := mp.getAllMetadata()
		if err != nil {
			return "", fmt.Errorf("failed to get metadata value by key %q because of error: %s", key, err.Error())
		}
	}

	toCamel := strnaming.NewCamel()
	toCamel.WithDelimiter('-')

	key = toCamel.Convert(key)
	path := strings.Split(key, "/")

	value, err := getValue(mp.cache, path)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata value by key %q because of error: %s", key, err.Error())
	}
	return value, nil
}

func (mp *CachingMetadataProvider) getAllMetadata() error {
	const (
		queryParam     = "?recursive=true"
		requestTimeout = 5 * time.Second
	)

	urlMetadata := getMetadataUrl() + queryParam

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlMetadata, nil)
	if err != nil {
		return fmt.Errorf("could not make request to get all metadata: %s", err.Error())
	}
	req.Header.Set("Metadata-Flavor", "Google")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("could not get all metadata: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to get all metadata returned status other than OK: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("response body returned by request to get all metadata read failed: %s", err.Error())
	}

	metadataStruct := new(structpb.Struct)
	err = metadataStruct.UnmarshalJSON(body)
	if err != nil {
		return fmt.Errorf("could not unmarshal response body returned by request to get all metadata: %s", err.Error())
	}

	mp.cache = metadataStruct
	return nil
}

func NewCachingMetadataProvider() MetadataProvider {
	return &CachingMetadataProvider{}
}
