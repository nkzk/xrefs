package k9s

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/nkzk/xtree/internal/utils"
)

func Install() error {
	dependencies := []string{"go", "k9s"}
	for _, dep := range dependencies {
		if !installed(dep) {
			return fmt.Errorf("%s is not installed", dep)
		}
	}

	output, err := utils.RunCommand("k9s", "info")
	if err != nil {
		return fmt.Errorf("failed to get k9s info: %w", err)
	}

	dir, err := getPluginDirectory(output)
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

func getPluginDirectory(input []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(input))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Plugins") {
			return strings.ReplaceAll(strings.Join(strings.Fields(scanner.Text())[1:], " "), " ", "\\ "), nil
		}
	}

	return "", errors.New("failed to get k9s plugind directory from k9s info")
}
