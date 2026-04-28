package k8s

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const (
	mockFluxKustomizationKind  = "Kustomization"
	mockXRKind                 = "MyXR"
	mockApplicationKind        = "Application"
	mockConfigMapKind          = "ConfigMap"
	mockClusterRoleBindingKind = "ClusterRoleBinding"
)

func mockFluxKustomization() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
			"kind":       mockFluxKustomizationKind,
			"metadata": map[string]any{
				"name":      "example",
				"namespace": "default",
			},
			"spec": map[string]any{},
			"status": map[string]any{
				"inventory": map[string]any{
					"entries": []any{
						map[string]any{
							"id": "default_example-cm__ConfigMap",
							"v":  "v1",
						},
						map[string]any{
							"id": "default_example-application_applications.azuread.m.upbound.io_Application",
							"v":  "v1beta1",
						},
						map[string]any{
							"id": "_example_rbac.authorization.k8s.io_ClusterRoleBinding",
							"v":  "v1",
						},
					},
				},
			},
		},
	}
}
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
						map[string]any{
							"apiVersion": "ayo",
							"kind":       "DoesNotExist",
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
			},
			"data": map[string]any{
				"key": "value",
			},
		},
	}
}

func mockClusterRoleBinding() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       mockClusterRoleBindingKind,
			"metadata": map[string]any{
				"name": "example",
			},
			"roleRef": map[string]any{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "ClusterRole",
				"name":     "crossplane-extras",
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
						"reason":             "ReconcileError",
						"observedGeneration": int64(7),
						"lastTransitionTime": "2025-10-10T12:55:42Z",
					},
				},
			},
		},
	}
}
