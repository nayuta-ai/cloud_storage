package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"testing"

	"gopkg.in/inf.v0"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func TestVpa(t *testing.T) {
	LOWER_LIMIT_CPU := inf.NewDec(100, 0)
	UPPER_LIMIT_CPU := inf.NewDec(300, 0)
	BASE_N := inf.NewDec(1000000000, 0) // 10**9
	LOWER_LIMIT_MEMORY := 300000000
	UPPER_LIMIT_MEMORY := 1000000000
	CONTAINER_NAME := "vpa-container"
	cpu, memory, err := fetchMetrics(CONTAINER_NAME)
	if err != nil {
		t.Errorf("%s", err)
	}
	//cpu := []int{100, 200, 300}
	//memory := []int{300, 400, 500}
	for i := 0; i < len(cpu); i++ {
		low := new(inf.Dec).QuoExact(LOWER_LIMIT_CPU, BASE_N)
		upper := new(inf.Dec).QuoExact(UPPER_LIMIT_CPU, BASE_N)
		if low.Cmp(cpu[i]) == 1 || upper.Cmp(cpu[i]) == -1 {
			t.Errorf("CPU resources is out of range.:%f", cpu[i])
		}
		if LOWER_LIMIT_MEMORY > memory[i] || memory[i] > UPPER_LIMIT_MEMORY {
			t.Errorf("Memory resources is out of range.:%d", memory[i])
		}
	}
}

func fetchMetrics(container_name string) ([]*inf.Dec, []int, error) {
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
	cpu_list := make([]*inf.Dec, 0)
	memory_list := make([]int, 0)
	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity := container.Usage.Cpu().AsDec()
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				return nil, nil, fmt.Errorf("error: Can't fetch memory resources")
			}
			if container.Name == container_name {
				cpu_list = append(cpu_list, cpuQuantity)
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
