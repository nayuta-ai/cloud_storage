package vpatest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/inf.v0"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalev1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/yaml"
)

// FetchInterface plays the role of making unit test easy.
type FetchInterface interface {
	FetchPodList(string, *Cluster) ([]corev1.Pod, error)
	FetchMetricsList(string, *Cluster) ([]*inf.Dec, []int64, error)
}

// Config plays the role of setting file for functions in recorder.go
type Config struct{}

// FetchPodList extracts PodList included containers information and so on.
func (c *Config) FetchPodList(namespace string, cluster *Cluster) ([]corev1.Pod, error) {
	if namespace == "" {
		namespace = "default"
	}
	podList, err := cluster.Client.Clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// FetchMetricsList extracts cpu and memory resource list included containers resource in the pod.
func (c *Config) FetchMetricsList(containerName string, cluster *Cluster) ([]*inf.Dec, []int64, error) {
	// connect to k8s cluster
	mc, err := metrics.NewForConfig(cluster.Config)
	if err != nil {
		return nil, nil, err
	}
	podMetrics, err := mc.MetricsV1beta1().PodMetricses("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	cpuList := make([]*inf.Dec, 0)
	memoryList := make([]int64, 0)
	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			cpuQuantity := container.Usage.Cpu().AsDec()
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				return nil, nil, fmt.Errorf("error: Can't fetch memory resources")
			}
			if container.Name == containerName {
				cpuList = append(cpuList, cpuQuantity)
				memoryList = append(memoryList, memQuantity)
			}
		}

	}
	return cpuList, memoryList, nil
}

// FetchDeploymentConfig extracts Deployment structure from YAML file.
func FetchDeploymentConfig(path string) (*appsv1.Deployment, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	deploymentConfig := &appsv1.Deployment{}
	err = yaml.Unmarshal(b, deploymentConfig)
	if err != nil {
		panic(err)
	}

	return deploymentConfig, nil
}

// FetchVPAConfig extracts VPA structure from YAML file.
func FetchVPAConfig(filepath string) (*autoscalev1.VerticalPodAutoscaler, error) {
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
