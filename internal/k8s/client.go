package k8s

import (
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client runtimeclient.Client
	// store
	resources *v1.APIResourceList // todo: storage-interface
}

func New() *Client {
	client, err := runtimeclient.New(getKubernetesConfig(), runtimeclient.Options{})
	if err != nil {
		log.Fatalf("failed to set up new runtime-client: %v", err)
	}

	return &Client{
		client: client,
	}
}

func (d *Client) isObjectNamespaced(obj runtime.Object) (bool, error) {
	return d.client.IsObjectNamespaced(obj)
}
