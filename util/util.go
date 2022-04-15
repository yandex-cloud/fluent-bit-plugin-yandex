package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fluent/fluent-bit-go/output"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func LevelFromString(level string) (logging.LogLevel_Level, error) {
	if v, ok := logging.LogLevel_Level_value[strings.ToUpper(level)]; ok {
		return logging.LogLevel_Level(v), nil
	}
	return logging.LogLevel_LEVEL_UNSPECIFIED, fmt.Errorf("bad level: %q", level)
}

func PayloadFromString(payload string) (*structpb.Struct, error) {
	result := new(structpb.Struct)
	err := result.UnmarshalJSON([]byte(payload))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getRecordValue(record map[interface{}]interface{}, path []string) (string, error) {
	var cur interface{} = record
	for _, p := range path {
		switch typed := cur.(type) {
		case map[interface{}]interface{}:
			cur = typed[p]
		case []interface{}:
			index, err := strconv.Atoi(p)
			if err != nil {
				return "", fmt.Errorf("incorrect path: expected number instead of %q", p)
			}
			if index >= len(typed) {
				return "", fmt.Errorf("incorrect path: index %q out of bound", p)
			}
			cur = typed[index]
		default:
			return "", errors.New("incorrect path")
		}
	}
	if cur == nil {
		return "", errors.New("incorrect path")
	}

	switch cur.(type) {
	case string, []byte:
		return ToString(cur), nil
	default:
		value, err := structpb.NewValue(Normalize(cur))
		if err != nil {
			return "", fmt.Errorf("failed to create protobuf value to marshal JSON: %s", err.Error())
		}
		content, err := value.MarshalJSON()
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %s", err.Error())
		}
		return string(content), nil
	}
}

func GetValue(from *structpb.Struct, path []string) (string, error) {
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
			if index >= len(cur.GetListValue().GetValues()) {
				return "", fmt.Errorf("incorrect path: index %q out of bound", p)
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

func ToString(raw interface{}) string {
	switch typed := raw.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func ToTime(raw interface{}) time.Time {
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

func Normalize(raw interface{}) interface{} {
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
			valSlice = append(valSlice, Normalize(el))
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
			if keyStr, ok := Normalize(key).(string); ok {
				valMap[keyStr] = Normalize(val)
			}
		}
		return valMap
	default:
		return raw
	}
}

// Truncate requires maxLen to be >= 3 (for '...')
func Truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
