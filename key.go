package main

import (
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	loggingpb "github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

type resourceKeys struct {
	resourceType string
	resourceID   string
}

func (rk *resourceKeys) logEntryResource() *loggingpb.LogEntryResource {
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
	resourceType *template
	resourceID   *template
}

// todo parse resourceKeys
func (pk *parseKeys) entry(ts time.Time, record map[interface{}]interface{}, tag string) (*loggingpb.IncomingLogEntry, resourceKeys) {
	var message string
	var level loggingpb.LogLevel_Level

	values := make(map[string]*structpb.Value)
	if len(pk.messageTag) > 0 {
		values[pk.messageTag] = structpb.NewStringValue(tag)
	}

	for k, v := range record {
		key, ok := k.(string)
		if !ok {
			continue
		}
		switch key {
		case pk.message:
			message = toString(v)
		case pk.level:
			levelName := toString(v)
			level, _ = levelFromString(levelName)
		default:
			value, err := structpb.NewValue(normalize(v))
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
	}, resourceKeys{} // todo fill
}
