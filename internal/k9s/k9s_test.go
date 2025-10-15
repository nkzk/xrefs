package k9s

import "testing"

func TestGetPluginDirectory(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "mac",
			input: `
 ____  __ ________       
|    |/  /   __   \______
|       /\____    /  ___/
|    \   \  /    /\___  \
|____|\__ \/____//____  /
         \/           \/ 

Version:           0.50.13
Config:            /Users/test/Library/Application Support/k9s/config.yaml
Custom Views:      /Users/test/Library/Application Support/k9s/views.yaml
Plugins:           /Users/test/Library/Application Support/k9s/plugins.yaml
Hotkeys:           /Users/test/Library/Application Support/k9s/hotkeys.yaml
Aliases:           /Users/test/Library/Application Support/k9s/aliases.yaml
Skins:             /Users/test/Library/Application Support/k9s/skins
Context Configs:   /Users/test/Library/Application Support/k9s/clusters
Logs:              /Users/test/Library/Application Support/k9s/k9s.log
Benchmarks:        /Users/test/Library/Application Support/k9s/benchmarks
ScreenDumps:       /Users/test/Library/Application Support/k9s/screen-dumps
`,
			want: "/Users/test/Library/Application\\ Support/k9s/plugins.yaml",
		},

		{
			name: "home",
			input: `
 ____  __ ________       
|    |/  /   __   \______
|       /\____    /  ___/
|    \   \  /    /\___  \
|____|\__ \/____//____  /
         \/           \/ 

Version:           0.50.13
Config:            /home/test/.k9s/config.yaml
Custom Views:      /home/test/.k9s/views.yaml
Plugins:           /home/test/.k9s/plugins.yaml
Hotkeys:           /home/test/.k9s/hotkeys.yaml
Aliases:           /home/test/.k9s/aliases.yaml
Skins:             /home/test/.k9s/skins
Context Configs:   /home/test/.k9s/clusters
Logs:              /home/test/.k9s/k9s.log
Benchmarks:        /home/test/.k9s/benchmarks
ScreenDumps:       /home/test/.k9s/screen-dumps
`,
			want: "/home/test/.k9s/plugins.yaml",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := getPluginDirectory([]byte(test.input))
			if err != nil {
				t.Fatalf("%v", err)
			}
			if got != test.want {
				t.Errorf("got %s want %s", got, test.want)
			}
		})
	}
}
