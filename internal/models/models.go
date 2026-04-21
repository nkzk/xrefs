package models

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resource struct {
	schema.GroupVersionKind
	Name      string
	Namespace *string

	ID       string
	Parent   *Resource
	Children []Resource
}
