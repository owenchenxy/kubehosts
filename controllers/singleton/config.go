package singleton

import (
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var insConfig *rest.Config
var onceConfig sync.Once

func GetConfig() *rest.Config {
	onceConfig.Do(func() {
		insConfig = getConfig()
	})
	return insConfig
}

func getConfig() *rest.Config {
	var home string
	home = os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	var restConfig *rest.Config
	var err error
	if restConfig, err = rest.InClusterConfig(); err != nil {
		if restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	return restConfig
}
