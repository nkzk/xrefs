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

	Expanded          bool // whether children are shown in the tree
	ChildrenLoaded    bool // whether children have been fetched from the API
	HasResourceRefs   bool // whether the resource has resourceRefs in its spec
	ResourceRefsCount int  // number of resourceRefs in the spec
	Loading           bool // placeholder item indicating children are loading

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
func (r Resource) FilterValue() string {
	if r.Unstructured == nil {
		return ""
	}
	val := r.Unstructured.GetKind() + " " + r.Unstructured.GetName()
	if r.Parent != nil && r.Parent.Unstructured != nil {
		val += " " + r.Parent.Unstructured.GetKind() + " " + r.Parent.Unstructured.GetName()
	}
	return val
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
