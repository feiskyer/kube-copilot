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
package tools

import (
	"os/exec"
	"strings"
)

// Trivy runs trivy against the image and returns the output
func Trivy(image string) (string, error) {
	image = strings.TrimSpace(image)
	if strings.HasPrefix(image, "trivy ") {
		image = strings.TrimPrefix(image, "trivy ")
	}

	if strings.HasPrefix(image, "image ") {
		image = strings.TrimPrefix(image, "image ")
	}

	cmd := exec.Command("trivy", "image", image, "--scanners", "vuln")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), err
	}

	return strings.TrimSpace(string(output)), nil
}
