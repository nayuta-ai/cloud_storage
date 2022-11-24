package vpatest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalev1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/yaml"
)

func (cluster *Cluster) fetchPodList(namespace string) ([]corev1.Pod, error) {
	if namespace == "" {
		namespace = "default"
	}
	podList, err := cluster.Client.Clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (cluster *Cluster) fetchMetricsList(containerName string) ([]*inf.Dec, []int64, error) {
	// connect to k8s cluster
	mc, err := metrics.NewForConfig(cluster.Config)
	if err != nil {
		return nil, nil, err
	}
	podMetrics, err := mc.MetricsV1beta1().PodMetricses("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	cpu_list := make([]*inf.Dec, 0)
	memory_list := make([]int64, 0)
	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity := container.Usage.Cpu().AsDec()
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				return nil, nil, fmt.Errorf("error: Can't fetch memory resources")
			}
			if container.Name == containerName {
				cpu_list = append(cpu_list, cpuQuantity)
				memory_list = append(memory_list, memQuantity)
			}
		}

	}
	return cpu_list, memory_list, nil
}

func fetchVPAConfig(filepath string) (*autoscalev1.VerticalPodAutoscaler, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	VPAConfig := &autoscalev1.VerticalPodAutoscaler{}
	err = yaml.Unmarshal(b, VPAConfig)
	if err != nil {
		return nil, err
	}

	return VPAConfig, nil
}
