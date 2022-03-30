package main

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
)

func TestGetRecordValue_Success(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b": "value",
		},
	}
	path := []string{"a", "b"}

	val, err := getRecordValue(record, path)

	assert.Nil(t, err)
	assert.Equal(t, "value", val)
}

func TestGetRecordValue_WithArray_Success(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": []interface{}{"value"},
	}
	path := []string{"a", "0"}

	val, err := getRecordValue(record, path)

	assert.Nil(t, err)
	assert.Equal(t, "value", val)
}

func TestGetRecordValue_JSONValue_Success(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b": map[interface{}]interface{}{
				"first":  "1st",
				"second": "2nd",
			},
		},
	}
	path := []string{"a", "b"}

	val, err := getRecordValue(record, path)

	assert.Nil(t, err)
	assert.Equal(t, "{\"first\":\"1st\",\"second\":\"2nd\"}", val)
}

func TestGetRecordValue_Fail(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": 123,
	}
	path := []string{"a", "b"}

	_, err := getRecordValue(record, path)

	assert.NotNil(t, err)
}

func TestGetRecordValue_WithArray_Fail(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": []interface{}{"value"},
	}
	path := []string{"a", "b"}

	_, err := getRecordValue(record, path)

	assert.NotNil(t, err)
}

func TestGetRecordValue_WithArray_OutOfBound_Fail(t *testing.T) {
	record := map[interface{}]interface{}{
		"a": []interface{}{"value"},
	}
	path := []string{"a", "1"}

	_, err := getRecordValue(record, path)

	assert.NotNil(t, err)
}

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
	expected := "{\"first\":\"1st\",\"second\":\"2nd\"}"
	assert.Equal(t, expected, val)
}

func TestGetValue_Fail(t *testing.T) {
	from := new(structpb.Struct)
	_ = from.UnmarshalJSON([]byte("{\"a\":123}"))
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
