package k9s

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type K9SConfig struct {
	Plugins []Plugin `json:"plugins" yaml:"plugins"`
}

type Plugin struct {
	Name        string
	ShortCut    string
	description string
	Command     string
	Background  bool
	Scopes      []string
}

func config(pluginKey, shortCut, command string) string {
	s := fmt.Sprintf(`
plugins:
  %s:
    shortCut: %s
    description: Show XR resourceRefs
    command: %s
    scopes:
    - "all"
    args:
    - --name 
    - $NAME
    - --namespace
    - $namespace
    - --resourceGroup
    - $RESOURCE_GROUP
    - --resourceName 
    - $RESOURCE_NAME
    - --resourceVersion
    - $RESOURCE__VERSION
	- --colComposition
    - $COL_COMPOSITION
    - --colCompositionRevision
    - $COL_COMPOSITION_REVISION
    background: false
    `, pluginKey, shortCut, command)

	return strings.TrimPrefix(s, "\n")
}
func CreatePluginFile(dstPath, pluginKey, shortCut, command string) error {
	if pluginKey == "" {
		return errors.New("plugin name cannot be empty")
	}
	if shortCut == "" {
		return errors.New("shortcut cannot be empty")
	}
	if command == "" {
		return errors.New("command cannot be empty")
	}

	if info, err := os.Stat(dstPath); err == nil {
		if info.Size() > 0 {
			return fmt.Errorf("%s already exists and is not empty; use append instead", dstPath)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	content := config(pluginKey, shortCut, command)

	return os.WriteFile(dstPath, []byte(content), 0o600)
}

func backupFile(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	name := filepath.Base(path)
	ts := time.Now().Format("20060102_150405")
	dstPath := filepath.Join("/tmp", fmt.Sprintf("%s.bak.%s", name, ts))

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	fmt.Printf("backed up old k9s config to %s\n", dstPath)

	return nil
}

func appendPlugin(doc []byte, key, shortcut, cmd, desc string, background bool, scopes []string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("plugin name (key) cannot be empty")
	}

	var root yaml.Node
	// If the file is empty or bad YAML, start a fresh doc: { plugins: {} }
	if len(bytes.TrimSpace(doc)) == 0 || yaml.Unmarshal(doc, &root) != nil {
		root = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Tag: "!!str", Value: "plugins"},
						{Kind: yaml.MappingNode}, // empty mapping
					},
				},
			},
		}
	} else if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		// Bad doc shape → reset to minimal
		root = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Tag: "!!str", Value: "plugins"},
						{Kind: yaml.MappingNode},
					},
				},
			},
		}
	}

	top := root.Content[0]
	if top.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("top-level not a mapping")
	}

	// Find or create "plugins" mapping
	var plugins *yaml.Node
	for i := 0; i < len(top.Content); i += 2 {
		k := top.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == "plugins" {
			plugins = top.Content[i+1]
			break
		}
	}
	if plugins == nil {
		plugins = &yaml.Node{Kind: yaml.MappingNode}
		top.Content = append(top.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "plugins"},
			plugins,
		)
	} else if plugins.Kind == yaml.ScalarNode && (plugins.Tag == "!!null" || plugins.Value == "") {
		// plugins: null → convert to mapping
		*plugins = yaml.Node{Kind: yaml.MappingNode}
	} else if plugins.Kind != yaml.MappingNode {
		return nil, fmt.Errorf(".plugins exists but is not a mapping")
	}

	// Duplicate check (exact key match)
	for i := 0; i < len(plugins.Content); i += 2 {
		k := plugins.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return nil, fmt.Errorf("plugin %q already exists", key)
		}
	}

	args := []string{
		"--name", "$NAME",
		"--namespace", "$namespace",
		"--resourceGroup", "$RESOURCE_GROUP",
		"--resourceName", "$RESOURCE_NAME",
		"--resourceversion", "$RESOURCE_VERSION",
		"--colComposition", "$COL_COMPOSITION",
		"--colCompositionRevision", "$COL_COMPOSITION_REVISION",
	}

	// Build value node
	val := &yaml.Node{Kind: yaml.MappingNode}
	appendKV(val, "shortCut", shortcut)
	appendKV(val, "description", desc)
	appendKV(val, "command", cmd)
	appendBool(val, "background", background)
	appendList(val, "scopes", scopes)
	appendList(val, "args", args)

	// Append `<key>: <val>` under plugins
	plugins.Content = append(plugins.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		val,
	)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func appendKV(m *yaml.Node, k, v string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v},
	)
}

func appendBool(m *yaml.Node, k string, b bool) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(b)},
	)
}

func appendList(m *yaml.Node, key string, values []string) {
	k := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range values {
		seq.Content = append(seq.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}
	m.Content = append(m.Content, k, seq)
}
