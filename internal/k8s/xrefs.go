package k8s

import (
	"context"

	"github.com/nkzk/xrefs/internal/models"
	v1 "k8s.io/api/core/v1"
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

	return result
}

// func (c K8sClient) ListResources(ctx context.Context, ref *v1.ObjectReference) *models.ResourceList {

// }
