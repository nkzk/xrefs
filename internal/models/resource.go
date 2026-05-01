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

	Expanded       bool // whether children are shown in the tree
	ChildrenLoaded bool // whether children have been fetched from the API

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
