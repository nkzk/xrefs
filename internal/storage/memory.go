package storage

import "github.com/nkzk/xrefs/internal/models"

type Memory struct {
	resources []models.Resource
}

func NewMemory() *Memory {
	return &Memory{}
}
