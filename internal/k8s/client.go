package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	GetUnstructured(ctx context.Context, ref *v1.ObjectReference) (*unstructured.Unstructured, error)
}

type K8sClient struct {
	client.Client // takes client.WithWatch
}

func NewK8sClient(client client.Client) *K8sClient {
	return &K8sClient{
		Client: client,
	}
}

// Gets the unstructured object for an objectreference
// Even if it fails, the unstructured object is filled with GVK and metadata.name and namespace, and the error is returned
func (c K8sClient) GetUnstructured(ctx context.Context, root *v1.ObjectReference) (*unstructured.Unstructured, error) {
	result := &unstructured.Unstructured{}
	result.SetGroupVersionKind(root.GroupVersionKind())

	err := c.Client.Get(
		ctx,
		client.ObjectKey{
			Name:      root.Name,
			Namespace: root.Namespace},
		result)
	// set name/namespace even if err, so the object is usable
	result.SetName(root.Name)
	result.SetNamespace(root.Namespace)

	return result, err
}

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c MockClient) GetUnstructured(ctx context.Context, r *v1.ObjectReference) (*unstructured.Unstructured, error) {
	switch r.Kind {
	case mockFluxKustomizationKind:
		return mockFluxKustomization(), nil
	case mockClusterRoleBindingKind:
		return mockClusterRoleBinding(), nil
	case mockXRKind:
		return mockXR(), nil
	case mockApplicationKind:
		return mockApplication(), nil
	case mockConfigMapKind:
		return mockConfigMap(), nil
	case mockRoleAssignmentKind:
		return mockRoleAssignment(), nil
	case mockUsageKind:
		switch r.Name {
		case "roleassignment-uses-providerconfig":
			return mockUsageRoleAssignmentUsesProviderConfig(), nil
		case "identity-uses-providerconfig":
			return mockUsageIdentityUsesProviderConfig(), nil
		}
		return mockUsage(), nil
	case mockUserAssignedIdentityKind:
		return mockUserAssignedIdentity(), nil
	case mockProviderConfigKind:
		if r.Name == "example" {
			return mockProviderConfig(), nil
		}
		return mockProviderConfig2(), nil
	default:
		return nil, fmt.Errorf("kind '%s' is not implemented in mock client: %w", r.Kind, apierrors.NewNotFound(
			schema.GroupResource{
				Group:    "",
				Resource: r.Kind,
			},
			r.Name,
		))
	}
}
