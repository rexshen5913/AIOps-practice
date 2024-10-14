package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	InCluster bool
	Command   string
	Kind      string
	KubePath  string
}

func main() {

	defer Recovery()
	runApp()
}

// 使用 pflag 解析參數
func parseArgs() (*Config, error) {
	useInCluster := pflag.Bool("incluster", false, "Use in-cluster config")
	kubeconfigPath := pflag.String("kubeconfig", "", "Absolute path to the kubeconfig file")

	// 在此解析 pflag 的參數
	pflag.Parse()

	// 取得剩餘的未解析位置參數，如 get <resource>
	args := pflag.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("usage: %s <command> <resource> [--incluster] [--kubeconfig=<path>]", os.Args[0])
	}

	command := args[0]
	kind := args[1]

	if command != "get" {
		return nil, fmt.Errorf("unsupported command: %s", command)
	}

	return &Config{
		InCluster: *useInCluster,
		Command:   command,
		Kind:      kind,
		KubePath:  *kubeconfigPath,
	}, nil
}

// 根據參數生成 Kubernetes 配置
func getK8sConfig(cfg *Config) (*rest.Config, error) {
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

func Recovery() {
	if r := recover(); r != nil {
		// unknown error
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("unknown error: %v", r)
		}
		trace := make([]byte, 4096)
		runtime.Stack(trace, true)
		log.Error().Fields(map[string]interface{}{
			"stack_trace": string(trace),
		}).Msg(err.Error())
	}
}

func runApp() {

	config, err := parseArgs()
	cfg := config
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing arguments")
		os.Exit(1)
	}

	k8sConfig, err := getK8sConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting Kubernetes config")
		os.Exit(1)
	}

	// dynamic client
	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Kubernetes clientset")

	}
	discoveryClient := clientset.Discovery()
	apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting API group resources")
	}

	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)

	gvk := schema.GroupVersionKind{
		Group:   "aiops.geektime.com",
		Version: "v1alpha1",
		Kind:    cfg.Kind,
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting REST mapping")
	}

	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resourceInterface := dynamicClient.Resource(mapping.Resource).Namespace("default")
	resources, err := resourceInterface.List(context, metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Error listing resources")
	}

	for _, resource := range resources.Items {
		fmt.Printf("Name: %s, Namespace: %s, UID: %s\n", resource.GetName(), resource.GetNamespace(), resource.GetUID())
	}

	// fmt.Printf("Successfully connected to Kubernetes. Command: %s, Resource: %s, InCluster: %v\n",
	// 	config.Command, config.Resource, config.InCluster)

}
