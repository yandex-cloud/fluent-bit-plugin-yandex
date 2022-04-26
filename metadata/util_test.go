package metadata

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/fluent-bit-plugin-yandex/v2/test"
)

func TestGetValue_Success(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":{\"b\":\"value\"}}"))
	path := []string{"a", "b"}

	val, err := getValue(from, path)

	assert.Nil(t, err)
	assert.Equal(t, "value", val)
}
func TestGetValue_WithArray_Success(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":[\"value\"]}"))
	path := []string{"a", "0"}

	val, err := getValue(from, path)

	assert.Nil(t, err)
	assert.Equal(t, "value", val)
}
func TestGetValue_JSONValue_Success(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":{\"b\":{\"first\":\"1st\",\"second\":\"2nd\"}}}"))
	path := []string{"a", "b"}

	val, err := getValue(from, path)

	assert.Nil(t, err)
	val = strings.ReplaceAll(val, " ", "") // hack
	assert.Equal(t, "{\"first\":\"1st\",\"second\":\"2nd\"}", val)
}
func TestGetValue_Fail(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":123}"))
	path := []string{"a", "b"}

	_, err := getValue(from, path)

	assert.NotNil(t, err)
}
func TestGetValue_NoSuchKey_Fail(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":{\"c\":\"value\"}}"))
	path := []string{"a", "b"}

	_, err := getValue(from, path)

	assert.NotNil(t, err)
}
func TestGetValue_WithArray_Fail(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":[\"value\"]}"))
	path := []string{"a", "b"}

	_, err := getValue(from, path)

	assert.NotNil(t, err)
}
func TestGetValue_WithArray_OutOfBound_Fail(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":[\"value\"]}"))
	path := []string{"a", "1"}

	_, err := getValue(from, path)

	assert.NotNil(t, err)
}

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
