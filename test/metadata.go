package test

import "fmt"

type MetadataProvider map[string]string

func (mp MetadataProvider) GetValue(key string) (string, error) {
	val, ok := mp[key]
	if ok {
		return val, nil
	}
	return "", fmt.Errorf("failed to get metadata value by key %q", key)
}
