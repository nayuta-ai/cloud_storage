package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func connectToKubernetes() (*rest.Config, *kubernetes.Clientset, error) {
	// The first step is to connect to the cluster.
	// Connection to the cluster requires K8s cluster config file, since we are connecting to an external remote cluster.
	// Since it is assumed that you have a cluster set-up, you must be having the config file ready.
	// It is assumed that the cluster config file is located at $HOME/.kube/config
	var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// The next step is to build the cluster config from url or config file path.
	// In this case, it builds the cluster config from config file path.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	// The final step is to create the clientset from config file.
	// It can get the some information from pods or nodes.
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, client, nil
}

func createJob(path string) {
	_, client, _ := connectToKubernetes()
	jobs := client.BatchV1().Jobs("default") // namespace: default
	jobConfig, _ := createJobStructure(path)
	_, err := jobs.Create(context.TODO(), jobConfig, metav1.CreateOptions{})
	if err != nil {
		log.Fatalln("Failed to create K8s job.")
	}
	log.Println("Created K8s job successfully")
}
