package k8s

import (
	"context"

	"github.com/nkzk/xrefs/internal/storage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceLoader interface {
	Load(name string, namespace *string)
}

type xrLoader struct {
	gvk     schema.GroupVersionKind
	client  dynamic.DynamicClient
	storage storage.Memory
}

func NewXRLoader(gvk schema.GroupVersionKind, client dynamic.DynamicClient, storage storage.Memory) *xrLoader {
	return &xrLoader{
		gvk:     gvk,
		client:  client,
		storage: storage,
	}
}

func (l xrLoader) Load(name string, namespace *string) {}

type fluxKustomizationLoader struct {
	gvk     schema.GroupVersionKind
	client  runtimeclient.Client
	storage storage.Memory
}

func NewFluxKustomizationLoader(client runtimeclient.Client, storage storage.Memory) *fluxKustomizationLoader {
	return &fluxKustomizationLoader{
		gvk: schema.GroupVersionKind{
			Group:   "kustomize.toolkit.fluxcd.io",
			Kind:    "Kustomization",
			Version: "v1",
		},
		client:  client,
		storage: storage,
	}
}

func (l fluxKustomizationLoader) Load(name string, namespace *string) {
	unstructured := unstructured.Unstructured{}
	l.client.Get(context.TODO(), runtimeclient.ObjectKey{Namespace: *namespace, Name: name}, &unstructured)
}
