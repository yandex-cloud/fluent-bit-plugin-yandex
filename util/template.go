package util

import (
	"fmt"
	"regexp"
	"strings"
)

type Template struct {
	format string
	keys   [][]string
}

var templateReg = regexp.MustCompile(`{[^{}]+}`)

func (t *Template) isTemplated() bool {
	return len(t.keys) != 0
}

func (t *Template) Parse(record map[interface{}]interface{}) (string, error) {
	if !t.isTemplated() {
		return t.format, nil
	}

	values := make([]interface{}, 0)
	for _, path := range t.keys {
		value, err := getRecordValue(record, path)
		if err != nil {
			return "", err
		}
		values = append(values, value)
	}

	return fmt.Sprintf(t.format, values...), nil
}

func NewTemplate(raw string) *Template {
	format := templateReg.ReplaceAllString(raw, "%s")
	paths := templateReg.FindAllString(raw, -1)

	keys := make([][]string, len(paths))
	for i, p := range paths {
		p = p[1 : len(p)-1]
		keys[i] = strings.Split(p, "/")
	}

	return &Template{
		format: format,
		keys:   keys,
	}
}
