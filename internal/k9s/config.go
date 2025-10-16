package k9s

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	return nil
}

func appendPlugin(doc []byte, key, shortcut, cmd, desc string, background bool, scopes []string) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(doc, &root); err != nil {
		return nil, err
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil, fmt.Errorf("bad yaml")
	}
	top := root.Content[0]
	if top.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("top-level not a mapping")
	}

	var plugins *yaml.Node
	for i := 0; i < len(top.Content); i += 2 {
		if top.Content[i].Value == "plugins" {
			plugins = top.Content[i+1]
			break
		}
	}
	if plugins == nil || plugins.Kind != yaml.MappingNode {
		return nil, fmt.Errorf(".plugins missing or not a mapping")
	}

	// refuse if key exists
	for i := 0; i < len(plugins.Content); i += 2 {
		if plugins.Content[i].Value == key {
			return nil, fmt.Errorf("plugin %q already exists", key)
		}
	}

	// append `<key>: {shortCut, description, command, background}`
	k := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	v := &yaml.Node{Kind: yaml.MappingNode}
	appendKV(v, "shortCut", shortcut)
	appendKV(v, "description", desc)
	appendKV(v, "command", cmd)
	appendList(v, "scopes", scopes)
	appendKV(v, "background", fmt.Sprintf("%t", background))

	plugins.Content = append(plugins.Content, k, v)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func appendKV(m *yaml.Node, k, v string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
		&yaml.Node{Kind: yaml.ScalarNode, Value: v}, // let yaml set tag
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
