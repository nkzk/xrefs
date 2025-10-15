package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type resourceRef struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
}

type XR struct {
	Metadata struct {
		Namespace string `json:"namespace" yaml:"namespace"`
	} `json:"metadata" yaml:"metadata"`
	Spec struct {
		Crossplane struct {
			ResourceRefs []resourceRef `json:"resourceRefs" yaml:"resourceRefs"`
		} `json:"crossplane" yaml:"crossplane"`
	} `json:"spec" yaml:"spec"`
}

func getRows(yamlString string) ([]row, error) {
	xr := &XR{}
	err := yaml.Unmarshal([]byte(yamlString), xr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xr: %w", err)
	}

	var result []row
	for _, resourceRef := range xr.Spec.Crossplane.ResourceRefs {
		result = append(result, row{
			Namespace:  xr.Metadata.Namespace,
			Kind:       resourceRef.Kind,
			ApiVersion: resourceRef.ApiVersion,
			Name:       resourceRef.Name,
			Synced:     "True",
			Ready:      "True",
		})
	}

	return result, nil
}

func main() {
	if _, err := tea.NewProgram(NewModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("failed to start: %v", err)
		os.Exit(1)
	}
}
