package k8s

import (
	"fmt"
	"path/filepath"

	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewK8sConfig(cfg *config.Config) (*rest.Config, error) {
	if cfg.InCluster {
		return rest.InClusterConfig()
	}

	// 如果沒有提供 kubeconfig 路徑，使用預設的 $HOME/.kube/config
	if cfg.KubePath == "" {
		if home := homedir.HomeDir(); home != "" {
			cfg.KubePath = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("cannot find kubeconfig file")
		}
	}

	return clientcmd.BuildConfigFromFlags("", cfg.KubePath)
}

// 初始化 k8s client
func NewK8sClients(config *rest.Config) (kubernetes.Interface, dynamic.Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating clientset: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating dynamic client: %v", err)
	}
	return clientset, dynamicClient, nil
}
