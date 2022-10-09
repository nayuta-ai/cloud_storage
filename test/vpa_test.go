package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func TestVpa(t *testing.T) {
	LOWER_LIMIT_CPU := 100
	UPPER_LIMIT_CPU := 300
	LOWER_LIMIT_MEMORY := 300
	UPPER_LIMIT_MEMORY := 1000
	CONTAINER_NAME := "vpa-container"
	cpu, memory, _ := fetchMetrics(CONTAINER_NAME)
	base := math.Pow10(6)
	for i := 0; i < len(cpu); i++ {
		if LOWER_LIMIT_CPU > cpu[i] || cpu[i] > UPPER_LIMIT_CPU {
			t.Errorf("CPU resources is out of range.")
		}
		if LOWER_LIMIT_MEMORY > memory[i]/int(base) || memory[i]/int(base) > UPPER_LIMIT_MEMORY {
			t.Errorf("Memory resources is out of range.")
		}
	}
}

func fetchMetrics(container_name string) ([]int, []int, error) {
	// The first step is to connect to the cluster.
	// Connection to the cluster requires K8s cluster config file, since we are connecting to an external remote cluster.
	// Since it is assumed that you have a cluster set-up, you must be having the config file ready.
	// It is assumed that the cluster config file is located at $HOME/.kube/config
	kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to the kubeconfig file to use")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	// connect to k8s cluster
	mc, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	podMetrics, err := mc.MetricsV1beta1().PodMetricses("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	cpu_list := make([]int, 0)
	memory_list := make([]int, 0)
	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity, ok := container.Usage.Cpu().AsInt64()
			if !ok {
				return nil, nil, fmt.Errorf("error: Can't fetch cpu resources")
			}
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				return nil, nil, fmt.Errorf("error: Can't fetch memory resources")
			}
			if container.Name == container_name {
				cpu_list = append(cpu_list, Int64ToInt(cpuQuantity))
				memory_list = append(memory_list, Int64ToInt(memQuantity))
			}
		}

	}
	return cpu_list, memory_list, nil
}

func Int64ToInt(i int64) int {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0
	} else {
		return int(i)
	}
}
