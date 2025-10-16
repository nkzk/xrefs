package config

import "os"

type Config struct {
	Name                   string
	Namespace              string
	ResourceName           string
	ResourceGroup          string
	ResourceVersion        string
	ColComposition         string
	ColCompositionRevision string
}

// k9s exports these environment variables for configuration
// ref https://k9scli.io/topics/plugins/
func Create() *Config {
	env := func(name string) string {
		val, exists := os.LookupEnv(name)
		if !exists {
			return ""
		}
		return val
	}

	return &Config{
		Name:                   env("NAME"),
		Namespace:              env("NAMESPACE"),
		ResourceName:           env("RESOURCE_NAME"),
		ResourceGroup:          env("RESOURCE_GROUP"),
		ResourceVersion:        env("RESOURCE_VERSION"),
		ColComposition:         env("COL_COMPOSITION"),
		ColCompositionRevision: env("COL_COMPOSITION_REVISION"),
	}
}

func (c Config) IsValid() bool {
	return c.Name != "" && c.Namespace != "" && c.ResourceName != "" && c.ResourceVersion != ""
}
