package yclient

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/model"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func levelFromString(level string) (logging.LogLevel_Level, error) {
	if v, ok := logging.LogLevel_Level_value[strings.ToUpper(level)]; ok {
		return logging.LogLevel_Level(v), nil
	}
	return logging.LogLevel_LEVEL_UNSPECIFIED, fmt.Errorf("bad level: %q", level)
}

func loggingWriteRequest(req *model.WriteRequest) *logging.WriteRequest {
	destination := &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: req.Destination.FolderID}}
	if len(req.Destination.LogGroupID) > 0 {
		destination = &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: req.Destination.LogGroupID}}
	}

	var resource *logging.LogEntryResource
	if len(req.Resource.Type) > 0 && len(req.Resource.ID) > 0 {
		resource = &logging.LogEntryResource{
			Type: req.Resource.Type,
			Id:   req.Resource.ID,
		}
	}

	entries := make([]*logging.IncomingLogEntry, 0)
	for _, entry := range req.Entries {
		level, _ := levelFromString(entry.Level)
		entries = append(entries, &logging.IncomingLogEntry{
			Level:       level,
			Message:     entry.Message,
			JsonPayload: entry.JSONPayload,
			Timestamp:   timestamppb.New(entry.Timestamp),
		})
	}

	level, _ := levelFromString(req.Defaults.Level)
	defaults := &logging.LogEntryDefaults{
		Level:       level,
		JsonPayload: req.Defaults.JSONPayload,
	}

	return &logging.WriteRequest{
		Destination: destination,
		Resource:    resource,
		Entries:     entries,
		Defaults:    defaults,
	}
}
