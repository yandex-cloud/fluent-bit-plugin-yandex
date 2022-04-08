package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
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
	GetValue(key string) (string, error)
}

type cachingMetadataProvider struct {
	mu         sync.RWMutex
	lastUpdate time.Time
	cache      *structpb.Struct
}

func (mp *cachingMetadataProvider) GetValue(key string) (string, error) {
	cache, err := mp.getAllMetadata()

	toCamel := strnaming.NewCamel()
	toCamel.WithDelimiter('-')
	key = toCamel.Convert(key)
	path := strings.Split(key, "/")

	value, err := getValue(cache, path)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata value by key %q because of error: %s", key, err.Error())
	}
	return value, nil
}

func (mp *cachingMetadataProvider) getAllMetadata() (*structpb.Struct, error) {
	const updateBackoff = time.Second
	mp.mu.RLock()
	passed := time.Since(mp.lastUpdate)
	if mp.cache != nil && passed < updateBackoff {
		defer mp.mu.RUnlock()
		return mp.cache, nil
	}
	mp.mu.RUnlock()

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

	mp.mu.Lock()
	mp.cache = metadataStruct
	mp.lastUpdate = time.Now()
	mp.mu.Unlock()
	return metadataStruct, nil
}

func NewCachingMetadataProvider() MetadataProvider {
	return &cachingMetadataProvider{}
}
