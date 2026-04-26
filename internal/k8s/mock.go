package k8s

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const (
	mockXRKind          = "MyXR"
	mockApplicationKind = "Application"
	mockConfigMapKind   = "ConfigMap"
)

func mockXR() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "my-group.io/v1alpha1",
			"kind":       mockXRKind,
			"metadata": map[string]any{
				"name":      "example",
				"namespace": "default",
			},
			"spec": map[string]any{
				"crossplane": map[string]any{
					"resourceRefs": []any{
						map[string]any{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"name":       "example",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "applications.azuread.m.upbound.io/v1beta1",
							"kind":       "Application",
							"name":       "example",
							"namespace":  "default",
						},
					},
				},
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Synced",
						"status":             "True",
						"reason":             "ReconcileSuccess",
						"observedGeneration": int64(7),
						"lastTransitionTime": "2025-10-10T12:55:42Z",
					},
					map[string]any{
						"type":               "Ready",
						"status":             "False",
						"reason":             "Creating",
						"observedGeneration": int64(7),
						"lastTransitionTime": "2025-10-10T12:55:42Z",
					},
				},
			},
		},
	}
}

func mockConfigMap() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       mockConfigMapKind,
			"metadata": map[string]any{
				"name":      "example-cm",
				"namespace": "default",
				"data": map[string]any{
					"key": "value",
				},
			},
		},
	}
}

func mockApplication() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "applications.azuread.m.upbound.io/v1beta1",
			"kind":       mockApplicationKind,
			"metadata": map[string]any{
				"name":      "example-application",
				"namespace": "default",
			},
			"spec": map[string]any{
				"crossplane": map[string]any{
					"forProvider": map[string]any{
						"displayName": "hi",
					},
				},
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Synced",
						"status":             "True",
						"reason":             "ReconcileSuccess",
						"observedGeneration": int64(7),
						"lastTransitionTime": "2025-10-10T12:55:42Z",
					},
					map[string]any{
						"type":               "Ready",
						"status":             "False",
						"reason":             "error: really long msg. the nodes are down cause my house burned down, contact the fire department, or try again in 15 minutes.",
						"observedGeneration": int64(7),
						"lastTransitionTime": "2025-10-10T12:55:42Z",
					},
				},
			},
		},
	}
}
