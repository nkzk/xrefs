package models

import (
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Resource struct {
	NotFound bool

	Ref          *v1.ObjectReference
	Unstructured *unstructured.Unstructured
	Conditions   Conditions

	ID       string
	Parent   *Resource
	Children []Resource

	Depth  int
	IsLast bool
	Prefix string

	Error error
}

func NewResource(
	parent *Resource,
	unstructured *unstructured.Unstructured,
	ref *v1.ObjectReference,
) *Resource {
	return &Resource{
		ID:           uuid.New().String(),
		Parent:       parent,
		Unstructured: unstructured,
		Ref:          ref,
	}
}

// implement tea list item interface

func (r Resource) Title() string       { return r.Unstructured.GetName() }
func (r Resource) Description() string { return r.Unstructured.GetNamespace() }
func (r Resource) FilterValue() string { return r.Unstructured.GetName() }

type Condition struct {
	Status             string `json:"status"`
	ConditionType      string `json:"type"`
	Reason             string `json:"reason"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

type Conditions []Condition

func (c Conditions) Get(t string) Condition {
	for _, cond := range c {
		if cond.ConditionType == t {
			return cond
		}
	}
	return Condition{}
}
