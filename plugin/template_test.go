package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplate_Success(t *testing.T) {
	raw := "begin_{simple}_{path/to/json/value}_end"

	templ := newTemplate(raw)

	format := "begin_%s_%s_end"
	keys := [][]string{
		{"simple"},
		{"path", "to", "json", "value"},
	}
	assert.NotNil(t, templ)
	assert.Equal(t, format, templ.format)
	assert.Equal(t, keys, templ.keys)
}

func TestParse_Success(t *testing.T) {
	templ := &template{
		format: "begin_%s_%s_end",
		keys: [][]string{
			{"simple"},
			{"path", "to"},
		},
	}
	record := map[interface{}]interface{}{
		"simple": "simple_value",
		"path": map[interface{}]interface{}{
			"to": "path_value",
		},
	}

	parsed, err := templ.parse(record)

	assert.Nil(t, err)
	assert.Equal(t, "begin_simple_value_path_value_end", parsed)
}
func TestParse_NotTemplated_Success(t *testing.T) {
	templ := &template{
		format: "begin_end",
		keys:   [][]string{},
	}
	record := map[interface{}]interface{}{}

	parsed, err := templ.parse(record)

	assert.Nil(t, err)
	assert.Equal(t, "begin_end", parsed)
}
func TestParse_Fail(t *testing.T) {
	templ := &template{
		format: "begin_%s_%s_end",
		keys: [][]string{
			{"simple"},
			{"path", "to"},
		},
	}
	record := map[interface{}]interface{}{
		"simple": "simple_value",
		"path":   "path_value",
	}

	_, err := templ.parse(record)

	assert.NotNil(t, err)
}
