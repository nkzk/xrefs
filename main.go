package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

type resourceRef struct {
	ApiVersion string `json:"apiVersion" yaml:"apiversion"`
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
}

type XR struct {
	Spec struct {
		Crossplane struct {
			ResourceRefs []resourceRef `json:"resourceRefs" yaml:"resourceRefs"`
		} `json:"crossplane" yaml:"crossplane"`
	} `json:"spec" yaml:"spec"`
}

func getResourceRefs(yamlString string) ([]resourceRef, error) {
	xr := &XR{}
	err := yaml.Unmarshal([]byte(yamlString), xr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xr: %w", err)
	}

	var result []resourceRef
	for _, resourceRef := range xr.Spec.Crossplane.ResourceRefs {
		result = append(result, resourceRef)
	}

	return result, nil
}

func main() {
	m := mock{}
	xrd := m.GetXRD()

	refs, err := getResourceRefs(xrd)
	if err != nil {
		log.Fatalf("failed to load xr: %v", err)
	}

	fmt.Printf("%+v\n", refs)
}
