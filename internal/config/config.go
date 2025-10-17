package config

type Config struct {
	Name                   string
	Namespace              string
	ResourceName           string
	ResourceGroup          string
	ResourceVersion        string
	ColComposition         string
	ColCompositionRevision string
}

func (c Config) IsValid() bool {
	return c.Name != "" && c.Namespace != "" && c.ResourceName != "" && c.ResourceVersion != "" && c.ColComposition != "" && c.ColCompositionRevision != ""
}
