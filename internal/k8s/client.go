package k8s

import (
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Storage interface{}

type ResourceLoader struct {
	inMemory bool

	client  runtimeclient.Client
	storage Storage
}

func New(inMemory bool, storage Storage) *ResourceLoader {
	client, err := runtimeclient.New(getKubernetesConfig(), runtimeclient.Options{})
	if err != nil {
		log.Fatalf("failed to set up new runtime-client: %v", err)
	}

	return &ResourceLoader{
		client:  client,
		storage: storage,
	}
}

func (d *ResourceLoader) isObjectNamespaced(obj runtime.Object) (bool, error) {
	return d.client.IsObjectNamespaced(obj)
}
