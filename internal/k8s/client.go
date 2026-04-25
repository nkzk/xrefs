package k8s

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setupclient(context string, diskCache bool) (clientcmd.ClientConfig, client.WithWatch, meta.RESTMapper, error) {
	cf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context},
	)

	kubeconfig, err := cf.ClientConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get kubeconfig: %v", err)
	}

	if kubeconfig.QPS == 0 {
		kubeconfig.QPS = 20
	}

	if kubeconfig.Burst == 0 {
		kubeconfig.Burst = 30
	}

	cl, err := client.NewWithWatch(kubeconfig, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to init kube client: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to set up discovery client: %v", err)
	}

	d := memory.NewMemCacheClient(discoveryClient)
	if diskCache {
		var err error
		d, err = disk.NewCachedDiscoveryClientForConfig(kubeconfig,
			getCacheDir(),
			getCacheDir(),
			10*time.Minute,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to set up discovery client with disk cache: %v", err)
		}
	}

	rmapper := restmapper.NewShortcutExpander(restmapper.NewDeferredDiscoveryRESTMapper(d), d, nil)

	return cf, cl, rmapper, nil
}

func getCacheDir() string {
	if cache := os.Getenv("XDG_CACHE_HOME"); cache != "" {
		return cache
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache")

}

// MappingFor returns the RESTMapping for the given resource or kind argument (TYPE[.VERSION][.GROUP][/NAME])
// Copied over from crossplane beta trace, which copied from cli-runtime pkg/resource Builder,
// https://github.com/kubernetes/cli-runtime/blob/9a91d944dd43186c52e0162e12b151b0e460354a/pkg/resource/builder.go#L768
func MappingFor(rmapper meta.RESTMapper, resourceOrKindArg string) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)

	gvk := schema.GroupVersionKind{}
	if fullySpecifiedGVR != nil {
		gvk, _ = rmapper.KindFor(*fullySpecifiedGVR)
	}

	if gvk.Empty() {
		gvk, _ = rmapper.KindFor(groupResource.WithVersion(""))
	}

	if !gvk.Empty() {
		return rmapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(resourceOrKindArg)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := rmapper.RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := rmapper.RESTMapping(groupKind, gvk.Version)
	if err != nil {
		// if we error out here, it is because we could not match a resource or a kind
		// for the given argument. To maintain consistency with previous behavior,
		// announce that a resource type could not be found.
		// if the error is _not_ a *meta.NoKindMatchError, then we had trouble doing discovery,
		// so we should return the original error since it may help a user diagnose what is actually wrong
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}

		return nil, err
	}

	return mapping, nil
}

func parseResourceName(resource, name string) (string, string, error) {
	if resource == "" {
		return "", "", errors.New("resource cannot be an empty string")
	}

	// Split the resource into its components
	splittedResource := strings.Split(resource, "/")
	length := len(splittedResource)

	if length == 1 {
		// Resource has only kind and the name is separately provided
		return splittedResource[0], name, nil
	}

	if length == 2 {
		// If a name is separately provided, error out
		if name == "" {
			return "", "", errors.New("invalid resource format, name cannot be defined twice, use TYPE[.VERSION][.GROUP][/NAME]")
		}

		// Resource includes both kind and name
		return splittedResource[0], splittedResource[1], nil
	}

	// Handle the case when resource format is invalid
	return "", "", errors.New("invalid resource format, use TYPE[.VERSION][.GROUP][/NAME]")
}

func resourceObjectFromMapping(mapping *meta.RESTMapping, clientconfig clientcmd.ClientConfig, name, namespace string) (*v1.ObjectReference, error) {
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if namespace == "" {
			var err error
			namespace, _, err = clientconfig.Namespace()
			if err != nil {
				return nil, fmt.Errorf("failed to get namespace from clientconfig")
			}
		}
	}

	return &v1.ObjectReference{
		Kind:       mapping.GroupVersionKind.Kind,
		APIVersion: mapping.GroupVersionKind.GroupVersion().String(),
		Name:       name,
		Namespace:  namespace,
	}, nil
}

// func (d *ResourceLoader) isObjectNamespaced(obj runtime.Object) (bool, error) {
// 	return d.client.IsObjectNamespaced(obj)
// }

// Resource{
// 				unstructured: unstructured.Unstructured{
// 					Object: map[string]any{
// 						"apiVersion": "v1",
// 						"kind":       "Deployment",
// 						"metadata": map[string]any{
// 							"name":      "parent-resource",
// 							"namespace": "default",
// 						},
// 					},
// 				},
// 				Children: []Resource{
// 					{
// 						unstructured: unstructured.Unstructured{
// 							Object: map[string]any{
// 								"apiVersion": "v1",
// 								"kind":       "ConfigMap",
// 								"metadata": map[string]any{
// 									"name":      "child-1",
// 									"namespace": "default",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						unstructured: unstructured.Unstructured{
// 							Object: map[string]any{
// 								"apiVersion": "v1",
// 								"kind":       "ConfigMap",
// 								"metadata": map[string]any{
// 									"name":      "child-2",
// 									"namespace": "default",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			}
