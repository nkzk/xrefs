package k8s

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const (
	mockFluxKustomizationKind    = "Kustomization"
	mockXRKind                   = "MyXR"
	mockApplicationKind          = "Application"
	mockConfigMapKind            = "ConfigMap"
	mockClusterRoleBindingKind   = "ClusterRoleBinding"
	mockUsageKind                = "Usage"
	mockUserAssignedIdentityKind = "UserAssignedIdentity"
	mockRoleAssignmentKind       = "RoleAssignment"
	mockProviderConfigUsageKind  = "ProviderConfigUsage"
	mockProviderConfigKind       = "ProviderConfig"
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
							"name":       "example-cm",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "applications.azuread.m.upbound.io/v1beta1",
							"kind":       "Application",
							"name":       "example-application",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "ayo",
							"kind":       "DoesNotExist",
							"name":       "example",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "protection.crossplane.io/v1beta1",
							"kind":       "Usage",
							"name":       "example-usage",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "example.io/v1beta1",
							"kind":       "UserAssignedIdentity",
							"name":       "example-identity",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "authorization.azure.m.upbound.io/v1beta1",
							"kind":       "RoleAssignment",
							"name":       "example-roleassignment",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "azure.m.upbound.io/v1beta1",
							"kind":       "ProviderConfig",
							"name":       "example-2",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "protection.crossplane.io/v1beta1",
							"kind":       "Usage",
							"name":       "roleassignment-uses-providerconfig",
							"namespace":  "default",
						},
						map[string]any{
							"apiVersion": "protection.crossplane.io/v1beta1",
							"kind":       "Usage",
							"name":       "identity-uses-providerconfig",
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

func mockUsage() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "protection.crossplane.io/v1beta1",
			"kind":       mockUsageKind,
			"metadata": map[string]any{
				"name":      "example-uses-providerconfig",
				"namespace": "default",
				"annotations": map[string]any{
					"crossplane.io/composition-resource-name": "roleassignment-uses-identity",
					"crossplane.io/usage-details":             "RoleAssignment/example-roleassignment uses UserAssignedIdentity/example-identity",
				},
				"labels": map[string]any{
					"crossplane.io/composite": "example",
				},
			},
			"spec": map[string]any{
				"by": map[string]any{
					"apiVersion": "example.io/v1beta1",
					"kind":       "RoleAssignment",
					"resourceRef": map[string]any{
						"name": "example-roleassignment",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"of": map[string]any{
					"apiVersion": "example.io/v1beta1",
					"kind":       "UserAssignedIdentity",
					"resourceRef": map[string]any{
						"name": "example-identity",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"replayDeletion": true,
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Ready",
						"status":             "True",
						"reason":             "Available",
						"observedGeneration": int64(3),
						"lastTransitionTime": "2026-03-19T09:24:26Z",
					},
				},
			},
		},
	}
}

func mockUsageIdentityUsesProviderConfig() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "protection.crossplane.io/v1beta1",
			"kind":       mockUsageKind,
			"metadata": map[string]any{
				"name":      "identity-uses-providerconfig",
				"namespace": "default",
				"annotations": map[string]any{
					"crossplane.io/composition-resource-name": "identity-uses-providerconfig",
					"crossplane.io/usage-details":             "identity/example uses ProviderConfig/example-2",
				},
			},
			"spec": map[string]any{
				"by": map[string]any{
					"apiVersion": "azure.m.upbound.io/v1beta1",
					"kind":       "UserAssignedIdentity",
					"resourceRef": map[string]any{
						"name": "example-identity",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"of": map[string]any{
					"apiVersion": "azure.m.upbound.io/v1beta1",
					"kind":       "ProviderConfig",
					"resourceRef": map[string]any{
						"name": "example-2",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"replayDeletion": true,
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Ready",
						"status":             "True",
						"reason":             "Available",
						"observedGeneration": int64(3),
						"lastTransitionTime": "2026-03-19T09:24:26Z",
					},
				},
			},
		},
	}
}
func mockUsageRoleAssignmentUsesProviderConfig() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "protection.crossplane.io/v1beta1",
			"kind":       mockUsageKind,
			"metadata": map[string]any{
				"name":      "roleassignment-uses-providerconfig",
				"namespace": "default",
				"annotations": map[string]any{
					"crossplane.io/composition-resource-name": "roleassignment-uses-identity",
					"crossplane.io/usage-details":             "RoleAssignment/example-roleassignment uses ProviderConfig/example-2",
				},
				"labels": map[string]any{
					"crossplane.io/composite": "example",
				},
			},
			"spec": map[string]any{
				"by": map[string]any{
					"apiVersion": "azure.m.upbound.io/v1beta1",
					"kind":       "RoleAssignment",
					"resourceRef": map[string]any{
						"name": "example-roleassignment",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"of": map[string]any{
					"apiVersion": "azure.m.upbound.io/v1beta1",
					"kind":       "ProviderConfig",
					"resourceRef": map[string]any{
						"name": "example-2",
					},
					"resourceSelector": map[string]any{
						"matchControllerRef": true,
					},
				},
				"replayDeletion": true,
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Ready",
						"status":             "True",
						"reason":             "Available",
						"observedGeneration": int64(3),
						"lastTransitionTime": "2026-03-19T09:24:26Z",
					},
				},
			},
		},
	}
}

func mockRoleAssignment() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "authorization.azure.m.upbound.io/v1beta1",
			"kind":       mockRoleAssignmentKind,
			"metadata": map[string]any{
				"name":      "example-roleassignment",
				"namespace": "default",
				"annotations": map[string]any{
					"crossplane.io/composition-resource-name": "roleassignment",
				},
				"labels": map[string]any{
					"crossplane.io/composite": "example",
					"crossplane.io/in-use":    "true",
					"identity":                "provider",
				},
				"ownerReferences": []any{
					map[string]any{
						"apiVersion":         "example.io/v1alpha1",
						"blockOwnerDeletion": true,
						"controller":         true,
						"kind":               "Environment",
						"name":               "example",
						"uid":                "00000000-0000-0000-0000-000000000001",
					},
				},
			},
			"spec": map[string]any{
				"forProvider": map[string]any{
					"name":               "example",
					"principalId":        "00000000-0000-0000-0000-000000000003",
					"principalType":      "ServicePrincipal",
					"roleDefinitionName": "Owner",
					"scope":              "/subscriptions/00000000-0000-0000-0000-000000000004",
				},
				"providerConfigRef": map[string]any{
					"name": "example-2",
					"kind": "ProviderConfig",
				},
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":               "Ready",
						"status":             "True",
						"reason":             "Available",
						"lastTransitionTime": "2026-03-19T09:27:49Z",
					},
					map[string]any{
						"type":               "Synced",
						"status":             "True",
						"reason":             "ReconcileSuccess",
						"observedGeneration": int64(3),
						"lastTransitionTime": "2026-04-30T15:21:47Z",
					},
				},
			},
		},
	}
}

func mockUserAssignedIdentity() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "managedidentity.azure.m.upbound.io/v1beta1",
			"kind":       mockUserAssignedIdentityKind,
			"metadata": map[string]any{
				"name":      "example-identity",
				"namespace": "default",
				"annotations": map[string]any{
					"crossplane.io/composition-resource-name": "identity",
				},
				"labels": map[string]any{
					"crossplane.io/composite": "example",
					"crossplane.io/in-use":    "true",
				},
				"ownerReferences": []any{
					map[string]any{
						"apiVersion":         "example.io/v1alpha1",
						"blockOwnerDeletion": true,
						"controller":         true,
						"kind":               "Environment",
						"name":               "example",
						"uid":                "00000000-0000-0000-0000-000000000001",
					},
				},
			},
			"spec": map[string]any{
				"forProvider": map[string]any{
					"location":          "norwayeast",
					"name":              "example-identity",
					"resourceGroupName": "example-rg",
				},
				"providerConfigRef": map[string]any{
					"name": "example-2",
					"kind": "ProviderConfig",
				},
			},
			"status": map[string]any{
				"atProvider": map[string]any{
					"clientId":          "00000000-0000-0000-0000-000000000005",
					"principalId":       "00000000-0000-0000-0000-000000000003",
					"location":          "norwayeast",
					"name":              "example-identity",
					"resourceGroupName": "example-rg",
				},
				"conditions": []any{
					map[string]any{
						"type":               "Ready",
						"status":             "True",
						"reason":             "Available",
						"lastTransitionTime": "2026-03-19T09:24:11Z",
					},
					map[string]any{
						"type":               "Synced",
						"status":             "True",
						"reason":             "ReconcileSuccess",
						"observedGeneration": int64(2),
						"lastTransitionTime": "2026-04-30T15:17:41Z",
					},
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
				"providerConfigRef": map[string]any{
					"name": "example-providerconfig",
					"kind": "ProviderConfig",
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

func mockProviderConfig() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "azure.m.upbound.io/v1beta1",
			"kind":       mockProviderConfigKind,
			"metadata": map[string]any{
				"name": "example-providerconfig",
			},
			"spec": map[string]any{},
		},
	}
}

func mockProviderConfig2() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "azure.m.upbound.io/v1beta1",
			"kind":       mockProviderConfigKind,
			"metadata": map[string]any{
				"name": "example-2",
			},
			"spec": map[string]any{},
		},
	}
}

func mockProviderConfigUsage() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "azure.m.upbound.io/v1beta1",
			"kind":       mockProviderConfigUsageKind,
			"metadata": map[string]any{
				"name": "example-providerconfigusage",
			},
			"providerConfigRef": map[string]any{
				"name": "example",
			},
			"resourceRef": map[string]any{
				"apiVersion": "azure.crossplane.io/v1alpha3",
				"kind":       "ResourceGroup",
				"name":       "example-rg",
				"uid":        "00000000-0000-0000-0000-000000000006",
			},
		},
	}
}
