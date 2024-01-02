package utils

import (
	"regexp"
	"strings"
)

// ExtractYaml extracts yaml from a markdown message.
func ExtractYaml(message string) string {
	r1 := regexp.MustCompile("(?s)```yaml(.*?)```")
	matches := r1.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) > 1 {
		return matches[1]
	}

	r2 := regexp.MustCompile("(?s)```(.*?)```")
	matches = r2.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
