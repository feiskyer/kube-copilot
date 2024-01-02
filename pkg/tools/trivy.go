package tools

import (
	"os/exec"
	"strings"
)

// Trivy runs trivy against the image and returns the output
func Trivy(image string) (string, error) {
	image = strings.TrimSpace(image)
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
