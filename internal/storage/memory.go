package storage

import "github.com/nkzk/xrefs/internal/models"

type Memory struct {
	resources []models.Xrefs
}

func NewMemory() *Memory {
	return &Memory{}
}
