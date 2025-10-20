package utils

import (
	"fmt"
	"os/exec"
)

func RunCommand(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed:%s %s\nerror: %w\noutput: %s", command, arg, err, output)
	}

	return output, nil
}
