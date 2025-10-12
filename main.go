package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

type resourceRef struct {
	ApiVersion string `json:"apiVersion" yaml:"apiversion"`
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
			namespace:  xr.Metadata.Namespace,
			kind:       resourceRef.Kind,
			apiVersion: resourceRef.ApiVersion,
			name:       resourceRef.Name,
			synced:     "True",
			ready:      "True",
		})
	}

	return result, nil
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Printf("failed to start: %v", err)
		os.Exit(1)
	}
}
