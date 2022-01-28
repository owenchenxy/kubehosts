package singleton

import (
	"sync"

	"k8s.io/client-go/kubernetes"
)

var insClient *kubernetes.Clientset
var onceClient sync.Once

func GetClient() *kubernetes.Clientset {
	onceClient.Do(func() {
		insClient = getClient()
	})
	return insClient
}

func getClient() *kubernetes.Clientset {
	config := GetConfig()
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}
