package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fluent/fluent-bit-go/output"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func levelFromString(level string) (logging.LogLevel_Level, error) {
	if v, ok := logging.LogLevel_Level_value[strings.ToUpper(level)]; ok {
		return logging.LogLevel_Level(v), nil
	}
	return logging.LogLevel_LEVEL_UNSPECIFIED, fmt.Errorf("bad level: %q", level)
}

func payloadFromString(payload string) (*structpb.Struct, error) {
	result := new(structpb.Struct)
	err := result.UnmarshalJSON([]byte(payload))
	if err != nil {
		return nil, err
	}
	return result, nil
}

var reg = regexp.MustCompile(`{{[^{}]+}}`)

func parsePayload(payload string) (string, error) {
	metadataCache, err := getAllMetadata()
	if err != nil {
		return "", err
	}
	var multierror error
	parsed := reg.ReplaceAllStringFunc(payload, func(t string) string {
		res, err := replaceTemplate(t, metadataCache)
		if err != nil {
			multierror = multierr.Append(multierror, err)
		}
		return res
	})
	return parsed, multierror
}

func replaceTemplate(t string, metadata *structpb.Struct) (string, error) {
	str := t[2 : len(t)-2]

	fields := strings.Split(str, ":")
	key := fields[0]
	defaultValue := ""
	if len(fields) >= 2 {
		defaultValue = fields[1]
	}

	metadataValue, err := getCachedMetadataValue(metadata, key)
	if err != nil {
		fmt.Printf("yc-logging: using default value %q for template %q because of error: %s\n", defaultValue, t, err.Error())
		return defaultValue, nil
	}
	return metadataValue, nil
}

func toString(raw interface{}) string {
	switch typed := raw.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func toTime(raw interface{}) time.Time {
	switch typed := raw.(type) {
	case output.FLBTime:
		return typed.Time
	case uint64:
		return time.Unix(int64(typed), 0)
	default:
		fmt.Println("time provided invalid, defaulting to now.")
		return time.Now()
	}
}

func normalize(raw interface{}) interface{} {
	switch typed := raw.(type) {
	case []byte:
		if utf8.Valid(typed) {
			return string(typed)
		}
		return typed
	case []interface{}:
		if len(typed) == 0 {
			return typed
		}
		valSlice := make([]interface{}, 0, len(typed))
		for _, el := range typed {
			valSlice = append(valSlice, normalize(el))
		}
		return valSlice
	case map[interface{}]interface{}:
		if len(typed) == 0 {
			if typed == nil {
				return nil
			}
			return map[string]interface{}{}
		}
		valMap := make(map[string]interface{}, len(typed))
		for key, val := range typed {
			if keyStr, ok := normalize(key).(string); ok {
				valMap[keyStr] = normalize(val)
			}
		}
		return valMap
	default:
		return raw
	}
}

// truncate requires maxLen to be >= 3 (for '...')
func truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
