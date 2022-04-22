package metadata

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

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

var metadataTemplateReg = regexp.MustCompile(`{{[^{}]+}}`)

func Parse(raw string, metadataProvider Provider) string {
	if ts := metadataTemplateReg.FindAllString(raw, -1); len(ts) == 0 {
		return raw
	}

	parsed := metadataTemplateReg.ReplaceAllStringFunc(raw, func(t string) string {
		return replaceTemplate(t, metadataProvider)
	})
	return parsed
}

func replaceTemplate(t string, metadataProvider Provider) string {
	str := t[2 : len(t)-2]

	fields := strings.Split(str, ":")
	key := fields[0]
	defaultValue := ""
	if len(fields) >= 2 {
		defaultValue = fields[1]
	}

	metadataValue, err := metadataProvider.GetValue(key)
	if err != nil {
		fmt.Printf("yc-logging: using default value %q for template %q because of error: %s\n", defaultValue, t, err.Error())
		return defaultValue
	}
	return metadataValue
}
