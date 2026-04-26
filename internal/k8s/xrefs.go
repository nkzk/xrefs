package k8s

import (
	"context"

	"github.com/nkzk/xrefs/internal/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	GetResource(ctx context.Context, ref *v1.ObjectReference) *models.Resource
	// ListResources(ctx context.Context, ref *v1.ObjectReference) *models.ResourceList
}

type K8sClient struct {
	client.Client // takes client.WithWatch
}

func NewK8sClient(client client.Client) *K8sClient {
	return &K8sClient{
		Client: client,
	}
}
func (c K8sClient) GetResource(ctx context.Context, root *v1.ObjectReference) *models.Resource {
	result := &models.Resource{}

	result.Ref = root
	result.Unstructured.SetGroupVersionKind(root.GroupVersionKind())

	err := c.Client.Get(
		ctx,
		client.ObjectKey{
			Name:      root.Name,
			Namespace: root.Namespace},
		&result.Unstructured)
	if err != nil {
		// set name/namespace anyway
		result.Unstructured.SetName(root.Name)
		result.Unstructured.SetNamespace(root.Namespace)
		result.Error = err
	}

	// // update conditions
	// conditions, ok, err := unstructured.NestedSlice(result.Unstructured.Object, "status", "conditions")
	// if ok && err == nil {
	// 	for _, c := range conditions {
	// 		condition, ok := c.(models.Condition)
	// 		if ok {
	// 			result.Conditions = append(result.Conditions, condition)
	// 		}
	// 	}
	// }

	return result
}

type MockClient struct{}

func (c MockClient) GetResource(ctx context.Context, root *v1.ObjectReference) *models.Resource {
	return &models.Resource{
		Parent: nil,
		ID:     "123",
		Unstructured: unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "alphav1",
				"kind":       "MyXR",
				"metadata": map[string]any{
					"name":      "example",
					"namespace": "default",
					"spec": map[string]any{
						"crossplane": map[string]any{
							"resourceRefs": []map[string]any{
								{
									"kind":       "Application",
									"apiVersion": "applications.azuread.m.upbound.io/v1beta1",
									"name":       "example",
								},
								{
									"kind":       "Secret",
									"apiVersion": "v1",
								},
							},
						},
					},
				},
				"status": map[string]any{
					"conditions": []map[string]any{
						{
							"type":               "Synced",
							"status":             "True",
							"reason":             "ReconcileSuccess",
							"observedGeneration": 7,
							"lastTransitionTime": "2025-10-10T12:55:42Z",
						},
						{
							"type":               "Ready",
							"status":             "False",
							"reason":             "Creating",
							"observedGeneration": 7,
							"lastTransitionTime": "2025-10-10T12:55:42Z",
						},
					},
				},
			},
		},
	}
}

// func (c K8sClient) ListResources(ctx context.Context, ref *v1.ObjectReference) *models.ResourceList {

// }
