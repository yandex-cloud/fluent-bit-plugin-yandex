package plugin

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
)

type parseKeys struct {
	level        string
	messageKeys  map[string]struct{}
	messageTag   string
	resourceType *template
	resourceID   *template
	streamName   *template
}

func (pk *parseKeys) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*model.Entry, model.Resource, error) {
	var message string
	var level string

	values := make(map[string]*structpb.Value)
	if len(pk.messageTag) > 0 {
		values[pk.messageTag] = structpb.NewStringValue(tag)
	}

	resourceType, err := pk.resourceType.parse(record)
	if err != nil {
		return nil, model.Resource{}, fmt.Errorf("failed to parse resource type: %s", err.Error())
	}
	resourceID, err := pk.resourceID.parse(record)
	if err != nil {
		return nil, model.Resource{}, fmt.Errorf("failed to parse resource ID: %s", err.Error())
	}
	streamName, err := pk.streamName.parse(record)
	if err != nil {
		return nil, model.Resource{}, fmt.Errorf("failed to parse stream name: %s", err.Error())
	}
	resource := model.Resource{
		Type: resourceType,
		ID:   resourceID,
	}

	for k, v := range record {
		key, ok := k.(string)
		if !ok {
			continue
		}

		if _, ok := pk.messageKeys[key]; ok {
			message = toString(v)
		} else {
			switch key {
			case pk.level:
				level = toString(v)
			default:
				value, err := structpb.NewValue(normalize(v))
				if err != nil {
					continue
				}
				values[key] = value
			}
		}
	}
	var payload *structpb.Struct
	if len(values) > 0 {
		payload = &structpb.Struct{
			Fields: values,
		}
	}
	return &model.Entry{
		Level:       level,
		StreamName:  streamName,
		Message:     message,
		JSONPayload: payload,
		Timestamp:   ts,
	}, resource, nil
}
