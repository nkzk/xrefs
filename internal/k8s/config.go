package k8s

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// tries to create a k8s *rest.config in this order (returning the first one that worked):
// 1) InClusterconfig
// 2) KUBECONFIG env
// 3) ~/.kube/config`
// Terminates the program if nothing worked
func getKubernetesConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("failed to set up kubeconfig: %v", err)
		}
	}

	return config
}
