package k8s

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"

	"github.com/nkzk/xrefs/internal/models"
)

type Discovery struct {
	client discovery.DiscoveryInterface

	// store
	resources *v1.APIResourceList
}

func NewDiscovery(client discovery.DiscoveryInterface) (*Discovery, error) {
	return &Discovery{
		client: client,
	}
}

// TODO: look into
// https://github.com/crossplane/crossplane/blob/d67e89cdad61ef2656334d0b249c0abcb4e56b26/internal/controller/apiextensions/composite/composition_functions.go#L487C26-L487C44

func (d *Discovery) isNamespaced(resource models.Resource) (bool, error) {
	gvk := resource.GroupVersionKind
	gv := gvk.GroupVersion().String()

	for _, r := range d.resources.APIResources {
		if r.Kind == gvk.Kind {
			return r.Namespaced, nil
		}
	}

	return false, fmt.Errorf("resource %q not found in %s", gvk.Kind, gv)
}
