package ui

import "testing"

func TestCreateKubectlCommand(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		group       string
		apiVersion  string
		resource    string
		namespace   string
		expected    string
		expectError bool
	}{
		{
			name:       "core v1 resource",
			kind:       "Secret",
			group:      "",
			apiVersion: "v1",
			resource:   "example",
			namespace:  "default",
			expected:   "kubectl get Secret.v1./example -n default -o yaml",
		},
		{
			name:       "resource with group in apiVersion (like in k8s yaml-output)",
			kind:       "RoleAssignment",
			group:      "",
			apiVersion: "applications.azuread.m.upbound.io/v1beta1",
			resource:   "example",
			namespace:  "default",
			expected:   "kubectl get RoleAssignment.applications.azuread.m.upbound.io/example -n default -o yaml",
		},
		{
			name:       "normal resource with group",
			kind:       "roleassignments",
			group:      "applications.azuread.m.upbound.io",
			apiVersion: "v1beta1",
			resource:   "example",
			namespace:  "default",
			expected:   "kubectl get roleassignments.v1beta1.applications.azuread.m.upbound.io/example -n default -o yaml",
		},
		{
			name:        "missing fields",
			kind:        "",
			apiVersion:  "v1",
			resource:    "example",
			namespace:   "default",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := CreateKubectlCommand(tt.kind, tt.group, tt.apiVersion, tt.resource, tt.namespace)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if cmd != tt.expected {
				t.Errorf("expected:\n  %s\ngot:\n  %s", tt.expected, cmd)
			}
		})
	}
}
