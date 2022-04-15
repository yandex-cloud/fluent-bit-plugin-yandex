package metadata

import (
	"fmt"
	"regexp"
	"strings"
)

var metadataTemplateReg = regexp.MustCompile(`{{[^{}]+}}`)

func Parse(raw string, metadataProvider MetadataProvider) string {
	if ts := metadataTemplateReg.FindAllString(raw, -1); len(ts) == 0 {
		return raw
	}

	parsed := metadataTemplateReg.ReplaceAllStringFunc(raw, func(t string) string {
		return replaceTemplate(t, metadataProvider)
	})
	return parsed
}

func replaceTemplate(t string, metadataProvider MetadataProvider) string {
	str := t[2 : len(t)-2]

	fields := strings.Split(str, ":")
	key := fields[0]
	defaultValue := ""
	if len(fields) >= 2 {
		defaultValue = fields[1]
	}

	metadataValue, err := metadataProvider.GetValue(key)
	if err != nil {
		fmt.Printf("yc-logging: using default value %q for template %q because of error: %s\n", defaultValue, t, err.Error())
		return defaultValue
	}
	return metadataValue
}
