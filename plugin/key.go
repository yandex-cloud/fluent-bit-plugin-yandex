package plugin

import (
	"fmt"
	"time"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/util"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	loggingpb "github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

type Resource struct {
	resourceType string
	resourceID   string
}

func (rk *Resource) LogEntryResource() *loggingpb.LogEntryResource {
	var resource *loggingpb.LogEntryResource
	if len(rk.resourceType) > 0 && len(rk.resourceID) > 0 {
		resource = &loggingpb.LogEntryResource{
			Type: rk.resourceType,
			Id:   rk.resourceID,
		}
	}
	return resource
}

type parseKeys struct {
	level        string
	message      string
	messageTag   string
	resourceType *util.Template
	resourceID   *util.Template
}

func (pk *parseKeys) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*loggingpb.IncomingLogEntry, Resource, error) {
	var message string
	var level loggingpb.LogLevel_Level

	values := make(map[string]*structpb.Value)
	if len(pk.messageTag) > 0 {
		values[pk.messageTag] = structpb.NewStringValue(tag)
	}

	resourceType, err := pk.resourceType.Parse(record)
	if err != nil {
		return nil, Resource{}, fmt.Errorf("failed to parse resource type: %s", err.Error())
	}
	resourceID, err := pk.resourceID.Parse(record)
	if err != nil {
		return nil, Resource{}, fmt.Errorf("failed to parse resource ID: %s", err.Error())
	}
	resource := Resource{
		resourceType: resourceType,
		resourceID:   resourceID,
	}

	for k, v := range record {
		key, ok := k.(string)
		if !ok {
			continue
		}
		switch key {
		case pk.message:
			message = util.ToString(v)
		case pk.level:
			levelName := util.ToString(v)
			level, _ = util.LevelFromString(levelName)
		default:
			value, err := structpb.NewValue(util.Normalize(v))
			if err != nil {
				continue
			}
			values[key] = value
		}
	}
	var payload *structpb.Struct
	if len(values) > 0 {
		payload = &structpb.Struct{
			Fields: values,
		}
	}
	return &loggingpb.IncomingLogEntry{
		Level:       level,
		Message:     message,
		JsonPayload: payload,
		Timestamp:   timestamppb.New(ts),
	}, resource, nil
}
