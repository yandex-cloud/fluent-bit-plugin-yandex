package main

import (
	"fmt"
	"regexp"
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

func parsePayload(payload *structpb.Struct) error {
	for _, v := range payload.GetFields() {
		err := parseTemplate(v)
		if err != nil {
			return err
		}
	}
	return nil
}

var reg = regexp.MustCompile(`{.*}`)

func parseTemplate(payloadValue *structpb.Value) error {
	switch value := payloadValue.AsInterface().(type) {
	case string:
		var err error
		parsed := reg.ReplaceAllStringFunc(value, func(t string) string {
			var res string
			res, err = replaceTemplate(t)
			return res
		})
		if err != nil {
			return err
		}
		*payloadValue = *structpb.NewStringValue(parsed)
	case map[string]interface{}:
		for _, v := range payloadValue.GetStructValue().GetFields() {
			err := parseTemplate(v)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for _, v := range payloadValue.GetListValue().GetValues() {
			err := parseTemplate(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceTemplate(t string) (string, error) {
	str := t[1 : len(t)-1]

	fields := strings.Split(str, ":")
	if len(fields) < 2 {
		return t, fmt.Errorf("configuration error: template %s must contain at least source and key diveded by \":\"", t)
	}
	source := fields[0]
	key := fields[1]
	defaultValue := ""
	if len(fields) >= 3 {
		defaultValue = fields[2]
	}

	switch source {
	case "metadata":
		metadataValue, err := getMetadataValue(key)
		if err != nil {
			return defaultValue, nil
		}
		return metadataValue, nil
	default:
		return t, nil
	}
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
