package tools

import (
	"os/exec"
	"strings"
)

// Kubectl runs the given kubectl command and returns the output.
func Kubectl(command string) (string, error) {
	if strings.HasPrefix(command, "kubectl") {
		command = strings.TrimSpace(strings.TrimPrefix(command, "kubectl"))
	}

	cmd := exec.Command("kubectl", strings.Split(command, " ")...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), err
	}

	return strings.TrimSpace(string(output)), nil
}
