package models

import (
	v1 "k8s.io/api/core/v1"
)

type RootDeletedMsg struct{}
type RootNotFoundMsg struct {
	Ref *v1.ObjectReference
}

type RootErrMsg struct {
	Err error
}

type QuitMsg struct{}

type ResourceList struct {
	Items []*Resource
	Error error
}
