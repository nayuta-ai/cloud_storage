package main

import (
    "fmt"
	"context"
	"flag"
	"os"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/tools/clientcmd"
    metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	// The first step is to connect to the cluster. 
	// Connection to the cluster requires K8s cluster config file, since we are connecting to an external remote cluster.
	// Since it is assumed that you have a cluster set-up, you must be having the config file ready.
	// It is assumed that the cluster config file is located at $HOME/.kube/config
    kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to the kubeconfig file to use")
    flag.Parse()
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig) 
    if err != nil {
        panic(err)
    }

	// connect to k8s cluster
    mc, err := metrics.NewForConfig(config)
    if err != nil {
        panic(err)
    }
    podMetrics, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    for _, podMetric := range podMetrics.Items {
        podContainers := podMetric.Containers
        for _, container := range podContainers {
            cpuQuantity, ok := container.Usage.Cpu().AsInt64()
            memQuantity, ok := container.Usage.Memory().AsInt64()
            if !ok {
                return
            }
            msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", container.Name, cpuQuantity, memQuantity)
            fmt.Println(msg)
        }

    }
}