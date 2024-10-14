package config

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type Config struct {
	InCluster bool
	Command   string
	Kind      string
	KubePath  string
}

func NewConfig() (*Config, error) {
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
