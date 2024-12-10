/*
Copyright 2023 - Present, Pengfei Ni

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
