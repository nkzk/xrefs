package ui

import (
	viewport "github.com/charmbracelet/bubbles/viewport"
	"github.com/nkzk/xrefs/internal/config"
)

type Viewport struct {
	viewport.Model

	cfg      config.Config
	updating bool
}

func (v Viewport) ID() string {
	return "viewport"
}

func (v Viewport) SetSize(w, h int) {
	v.Model.Width = max(0, w-2)
	v.Model.Height = max(0, h-2)
}
