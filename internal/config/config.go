package config

import "io"

type Config struct {
	Name                   string
	Namespace              string
	ResourceName           string
	ResourceGroup          string
	ResourceVersion        string
	ColComposition         string
	ColCompositionRevision string
	Mock                   bool

	Debug       bool
	DebugWriter io.Writer
	DebugPath   string
}

func (c Config) IsValid() bool {
	return c.Name != "" && c.Namespace != "" && c.ResourceName != "" && c.ResourceVersion != "" && c.ColComposition != "" && c.ColCompositionRevision != ""
}
