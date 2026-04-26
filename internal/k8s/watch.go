package k8s

import (
	"context"
	"errors"
	"time"

	"github.com/nkzk/xrefs/internal/models"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesResourceWatcher struct {
	Client client.Client
}

func NewKubernetesResourceWatcher(client client.Client) *KubernetesResourceWatcher {
	return &KubernetesResourceWatcher{
		Client: client,
	}
}

func (w KubernetesResourceWatcher) WatchResource(
	ctx context.Context,
	root *models.Resource,
) (watch.Interface, error) {
	opts := &client.ListOptions{
		Namespace:     root.Unstructured.GetNamespace(),
		FieldSelector: fields.OneTermEqualSelector("metadata.name", root.Unstructured.GetName()),
	}

	obj := root.Unstructured.DeepCopy()

	watchClient, ok := w.Client.(client.WithWatch)
	if !ok {
		return nil, errors.New("client does not support watch")
	}

	return watchClient.Watch(ctx, obj, opts)
}

type MockResourceWatcher struct{}

func NewMockResourceWatcher() *MockResourceWatcher {
	return &MockResourceWatcher{}
}

func (w MockResourceWatcher) WatchResource(
	ctx context.Context,
	root *models.Resource,
) (watch.Interface, error) {
	fw := watch.NewFake()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fw.Modify(root.Unstructured.DeepCopy())

			case <-ctx.Done():
				fw.Stop()
				return
			}
		}
	}()

	return fw, nil
}
