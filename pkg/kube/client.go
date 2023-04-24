package kube

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var DefaultConfigPath = os.Getenv("KUBECONFIG")

func NewClient(kubeconfig string) (*kubernetes.Clientset, error) {
	var (
		err        error
		restConfig *rest.Config
	)

	if kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("error build config from flags / %w", err)
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error bind in-cluster config / %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error create new clientset / %w", err)
	}

	return clientset, nil

}
