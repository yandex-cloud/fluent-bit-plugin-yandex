package yclient

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"

	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/model"
)

func levelFromString(level string) (logging.LogLevel_Level, error) {
	if v, ok := logging.LogLevel_Level_value[strings.ToUpper(level)]; ok {
		return logging.LogLevel_Level(v), nil
	}
	return logging.LogLevel_LEVEL_UNSPECIFIED, fmt.Errorf("bad level: %q", level)
}

func loggingDestination(from *model.Destination) *logging.Destination {
	destination := &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: from.FolderID}}
	if len(from.LogGroupID) > 0 {
		destination = &logging.Destination{Destination: &logging.Destination_LogGroupId{LogGroupId: from.LogGroupID}}
	}
	return destination
}

func logEntryDefaults(from *model.Defaults) (*logging.LogEntryDefaults, error) {
	if from == nil {
		return nil, nil
	}

	defaults := new(logging.LogEntryDefaults)

	if len(from.Level) > 0 {
		level, err := levelFromString(from.Level)
		if err != nil {
			return nil, err
		}
		defaults.Level = level
		fmt.Printf("yc-logging: will use %s as default level\n", level.String())
	}

	if from.JSONPayload != nil {
		defaults.JsonPayload = from.JSONPayload
		data, _ := from.JSONPayload.MarshalJSON()
		if data != nil {
			fmt.Printf("yc-logging: will default payload:\n%s\n", string(data))
		}
	}

	return defaults, nil
}
