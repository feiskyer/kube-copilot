package tools

import (
	"os/exec"
	"strings"
)

// PythonREPL runs the given Python script and returns the output.
func PythonREPL(script string) (string, error) {
	cmd := exec.Command("python3", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), err
	}

	return strings.TrimSpace(string(output)), nil
}
