package plugin

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/fluent/fluent-bit-go/output"
	"google.golang.org/protobuf/types/known/structpb"
)

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
			return "", fmt.Errorf("incorrect path")
		}
	}
	if cur == nil {
		return "", errors.New("incorrect path")
	}

	switch cur.(type) {
	case string, []byte:
		return toString(cur), nil
	default:
		value, err := structpb.NewValue(normalize(cur))
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
	case []interface{}:
		if dt, ok := typed[0].(output.FLBTime); ok {
			return dt.Time
		}
		fmt.Printf("provided time (%+v) invalid: defaulting to now.\n", typed)
		return time.Now()
	default:
		fmt.Printf("provided time (%+v) invalid: defaulting to now.\n", typed)
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
