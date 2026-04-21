package k8s

import (
	"log"

	"k8s.io/client-go/dynamic"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newRuntimeClient(inMemory bool) runtimeclient.Client {
	var client runtimeclient.Client
	if inMemory {
		client = fakectrlruntimeclient.NewFakeClient()
	} else {
		var err error
		client, err = runtimeclient.New(
			getKubernetesConfig(),
			runtimeclient.Options{})
		if err != nil {
			log.Fatalf("failed to set up new runtime-client: %v", err)
		}
	}

	return client
}

func newDynamicClient() *dynamic.DynamicClient {
	config := getKubernetesConfig()
	return dynamic.NewForConfigOrDie(config)
}

// func (d *ResourceLoader) isObjectNamespaced(obj runtime.Object) (bool, error) {
// 	return d.client.IsObjectNamespaced(obj)
// }
