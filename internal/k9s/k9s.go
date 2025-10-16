package k9s

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nkzk/xtree/internal/utils"
	"gopkg.in/yaml.v3"
)

func Install(shortcut string) error {
	const (
		pluginKey   = "xrefs"
		description = "Show XR resourceRefs"
	)

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

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

	pluginPath, err := getPluginDirectory(output)
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	background := false
	scopes := []string{"all"}

	if !fileExists(pluginPath) {
		if err := os.MkdirAll(filepath.Dir(pluginPath), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", pluginPath, err)
		}

		root := yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{Kind: yaml.MappingNode},
			},
		}

		top := root.Content[0]

		top.Content = append(top.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "plugins"},
			&yaml.Node{Kind: yaml.MappingNode},
		)

		bin, err := yaml.Marshal(&root)
		if err != nil {
			return fmt.Errorf("marshal seed yaml: %w", err)
		}

		out, err := appendPlugin(bin, pluginKey, shortcut, executablePath, description, background, scopes)
		if err != nil {
			return fmt.Errorf("append plugin into new file: %w", err)
		}

		if err := os.WriteFile(pluginPath, out, 0o644); err != nil {
			return fmt.Errorf("write new plugins file: %w", err)
		}

		return nil
	}

	if err := backupFile(pluginPath); err != nil {
		return fmt.Errorf("failed to backup file: %w", err)
	}

	in, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", pluginPath, err)
	}

	if len(in) == 0 {
		in = []byte("plugins: {}\n")
	}

	out, err := appendPlugin(in, pluginKey, shortcut, executablePath, description, background, scopes)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("plugin %q already present", pluginKey)
		}
		return fmt.Errorf("append plugin: %w", err)
	}

	if err := os.WriteFile(pluginPath, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", pluginPath, err)
	}

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
		line := scanner.Text()
		if strings.Contains(line, "Plugins") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				return strings.Join(fields[1:], " "), nil
			}
			break
		}
	}
	return "", errors.New("failed to get k9s plugin directory from k9s info")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}
