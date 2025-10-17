package utils

import (
	"fmt"
	"os/exec"
)

func RunCommand(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get command output: %w, %s", err, output)
	}

	return output, nil
}
