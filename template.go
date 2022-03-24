package main

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

type template struct {
	format string
	keys   [][]string
}

var templateReg = regexp.MustCompile(`{[^{}]+}`)

func (t *template) isTemplated() bool {
	return len(t.keys) != 0
}

func (t *template) parse(payload *structpb.Struct) (string, error) {
	if !t.isTemplated() {
		return t.format, nil
	}

	values := make([]interface{}, 0)
	for _, path := range t.keys {
		value, err := getValue(payload, path)
		if err != nil {
			return "", fmt.Errorf("failed to parse template because of error: %s", err.Error())
		}
		values = append(values, value)
	}

	return fmt.Sprintf(t.format, values...), nil
}

func newTemplate(raw string) *template {
	format := templateReg.ReplaceAllString(raw, "%s")
	paths := templateReg.FindAllString(raw, -1)

	keys := make([][]string, len(paths))
	for i, p := range paths {
		p = p[1 : len(p)-1]
		keys[i] = strings.Split(p, "/")
	}

	return &template{
		format: format,
		keys:   keys,
	}
}