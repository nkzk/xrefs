package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func Install() error {
	dependencies := []string{"go", "k9s"}
	for _, dep := range dependencies {
		if !installed(dep) {
			return fmt.Errorf("%s is not installed", dep)
		}
	}

	dir, err := getPluginDirectory()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	fmt.Printf("%s", dir)

	return nil
}

func installed(s string) bool {
	_, err := exec.LookPath(s)
	if err != nil {
		log.Printf("error: failed to look up command %s", err)
		return false
	}
	return true
}

func run(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get command output: %w", err)
	}

	return output, nil
}

func getPluginDirectory() (string, error) {
	output, err := run("k9s", "info")
	if err != nil {
		return "", fmt.Errorf("failed to get k9s info: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Plugins") {
			return scanner.Text(), nil
		}
	}

	return "", errors.New("failed to get k9s plugind directory from k9s info")
}
