package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntry_Success(t *testing.T) {
	pk := parseKeys{
		level:        "level",
		message:      "message",
		messageTag:   "tag_key",
		resourceType: newTemplate("resource_type"),
		resourceID:   newTemplate("resource_id"),
	}
	ts := time.Now()
	record := map[interface{}]interface{}{
		"level":   "INFO",
		"message": "record_message",
	}
	tag := "tag"

	entry, res, err := pk.entry(ts, record, tag)

	assert.Nil(t, err)
	assert.Equal(t, Resource{resourceType: "resource_type", resourceID: "resource_id"}, res)
	assert.Equal(t, logging.LogLevel_INFO, entry.Level)
	assert.Equal(t, "record_message", entry.Message)
	assert.Equal(t, timestamppb.New(ts), entry.Timestamp)
	tagVal, ok := entry.JsonPayload.AsMap()["tag_key"]
	assert.True(t, ok)
	assert.Equal(t, "tag", tagVal)
}
func TestEntry_TemplatedResource_Success(t *testing.T) {
	pk := parseKeys{
		resourceType: newTemplate("{simple}"),
		resourceID:   newTemplate("resource_{json/path}"),
	}
	ts := time.Now()
	record := map[interface{}]interface{}{
		"simple": "resource_type",
		"json": map[interface{}]interface{}{
			"path": "id",
		},
	}

	_, res, err := pk.entry(ts, record, "")

	assert.Nil(t, err)
	assert.Equal(t, Resource{resourceType: "resource_type", resourceID: "resource_id"}, res)
}
func TestEntry_TemplatedResourceID_Fail(t *testing.T) {
	pk := parseKeys{
		resourceType: newTemplate("{simple}"),
		resourceID:   newTemplate("resource_{json/path}"),
	}
	ts := time.Now()
	record := map[interface{}]interface{}{
		"simple": "resource_type",
	}

	_, _, err := pk.entry(ts, record, "")

	assert.NotNil(t, err)
}
func TestEntry_TemplatedResourceType_Fail(t *testing.T) {
	pk := parseKeys{
		resourceType: newTemplate("{simple}"),
		resourceID:   newTemplate("resource_{json/path}"),
	}
	ts := time.Now()
	record := map[interface{}]interface{}{
		"json": map[interface{}]interface{}{
			"path": "id",
		},
	}

	_, _, err := pk.entry(ts, record, "")

	assert.NotNil(t, err)
}
