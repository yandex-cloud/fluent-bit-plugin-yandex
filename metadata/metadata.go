package metadata

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/startdusk/strnaming"
	"google.golang.org/protobuf/types/known/structpb"
)

func getMetadataURL(instanceMetadataAddr string) (*url.URL, error) {
	const (
		keyMetadataURLEnv = "YC_METADATA_URL"
		urlPath           = "/computeMetadata/v1/"
	)
	metadataEndpoint := os.Getenv(keyMetadataURLEnv)
	if len(metadataEndpoint) == 0 {
		metadataEndpoint = "http://" + instanceMetadataAddr
	}
	urlMetadata, err := url.Parse(metadataEndpoint)
	if err != nil {
		return nil, err
	}
	urlMetadata.Path = urlPath
	return urlMetadata, nil
}

type Provider interface {
	GetValue(key string) (string, error)
}

type cachingProvider struct {
	mu                   sync.RWMutex
	instanceMetadataAddr string
	lastUpdate           time.Time
	cache                *structpb.Struct
}

func (mp *cachingProvider) GetValue(key string) (string, error) {
	cache, err := mp.getAllMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to get metadata value by key%q: %s", key, err.Error())
	}

	toCamel := strnaming.NewCamel()
	toCamel.WithDelimiter('-')
	key = toCamel.Convert(key)
	path := strings.Split(key, "/")

	value, err := getValue(cache, path)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata value by key %q: %s", key, err.Error())
	}
	return value, nil
}

func (mp *cachingProvider) getAllMetadata() (*structpb.Struct, error) {
	const (
		updateBackoff   = time.Second
		queryParamKey   = "recursive"
		queryParamValue = "true"
		requestTimeout  = 5 * time.Second
	)

	mp.mu.RLock()

	passed := time.Since(mp.lastUpdate)
	if mp.cache != nil && passed < updateBackoff {
		defer mp.mu.RUnlock()
		return mp.cache, nil
	}
	urlMetadata, err := getMetadataURL(mp.instanceMetadataAddr)
	if err != nil {
		return nil, fmt.Errorf("incorrect metadata URL: %s", err.Error())
	}
	q := urlMetadata.Query()
	q.Set(queryParamKey, queryParamValue)
	urlMetadata.RawQuery = q.Encode()

	mp.mu.RUnlock()

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlMetadata.String(), nil)
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

func NewCachingProvider(instanceMetadataAddr string) Provider {
	return &cachingProvider{
		instanceMetadataAddr: instanceMetadataAddr,
	}
}
