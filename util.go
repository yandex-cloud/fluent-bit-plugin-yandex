package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fluent/fluent-bit-go/output"
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

func getValue(from *structpb.Struct, path []string) (string, error) {
	cur := structpb.NewStructValue(from)
	for _, p := range path {
		switch cur.GetKind().(type) {
		case *structpb.Value_StructValue:
			cur = cur.GetStructValue().GetFields()[p]
		case *structpb.Value_ListValue:
			index, err := strconv.Atoi(p)
			if err != nil {
				return "", fmt.Errorf("incorrect path: expected number instead of %q", p)
			}
			cur = cur.GetListValue().GetValues()[index]
		default:
			return "", errors.New("incorrect path")
		}
	}
	if cur == nil {
		return "", errors.New("incorrect path")
	}

	if _, ok := cur.GetKind().(*structpb.Value_StringValue); ok {
		return cur.GetStringValue(), nil
	}

	content, err := cur.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %s", err.Error())
	}
	return string(content), nil
}

var metadataTemplateReg = regexp.MustCompile(`{{[^{}]+}}`)
var metadataCache *structpb.Struct

func parseWithMetadata(raw string) (string, error) {
	if ts := metadataTemplateReg.FindAllString(raw, -1); len(ts) == 0 {
		return raw, nil
	}

	var err error
	if metadataCache == nil {
		metadataCache, err = getAllMetadata()
	}
	if err != nil {
		return "", err
	}

	parsed := metadataTemplateReg.ReplaceAllStringFunc(raw, func(t string) string {
		return replaceTemplate(t, metadataCache)
	})
	return parsed, nil
}

func replaceTemplate(t string, metadata *structpb.Struct) string {
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
		return defaultValue
	}
	return metadataValue
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
