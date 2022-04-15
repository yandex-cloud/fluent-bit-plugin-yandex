package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/test"
)

func TestParse_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{
		"key": "val",
	}
	raw := "begin_{{key}}"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "begin_val", parsed)
}
func TestParse_JSONResult_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{
		"key": "{\"first\":\"1st\",\"second\":\"2nd\"}",
	}
	raw := "{\"key\":{{key}}}"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "{\"key\":{\"first\":\"1st\",\"second\":\"2nd\"}}", parsed)
}
func TestParse_SimpleJsonPayloadTemplate_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{
		"key": "val",
	}
	raw := "{from/json/payload}_{{key}}"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "{from/json/payload}_val", parsed)
}
func TestParse_NestedJsonPayloadTemplate_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{
		"key": "val",
	}
	raw := "{from/json/payload/{{key}}}"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "{from/json/payload/val}", parsed)
}
func TestParse_DefaultValue_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{}
	raw := "{{key}}_end"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "_end", parsed)
}
func TestParse_DefinedDefaultValue_Success(t *testing.T) {
	var metadataProvider = test.MetadataProvider{}
	raw := "{{key:default}}_end"

	parsed := Parse(raw, metadataProvider)

	assert.Equal(t, "default_end", parsed)
}
