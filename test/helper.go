package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/inf.v0"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func connectToKubernetes() (*rest.Config, *clientset.Clientset, error) {
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
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, client, nil
}

// It can fetch pod lists in the cluster.
func fetchPodList(clientset *clientset.Clientset) ([]corev1.Pod, error) {
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// It can fetch container lists in the pod.
func fetchContainerlist(pod corev1.Pod) []corev1.Container {
	var containerList []corev1.Container
	containerList = append(containerList, pod.Spec.Containers...)
	return containerList
}

// Check the change which VPA affects on the request resources for each pod.
func check_vpa(container corev1.Container) bool {
	return fmt.Sprintf("%v", container.Resources.Requests.Memory()) == fmt.Sprintf("%v", resource.NewQuantity(62914560, resource.BinarySI))
}

// Fetch all resource information such as CPU, memory for each pod.
func fetchMetrics(config *rest.Config, container_name string) ([]*inf.Dec, []int, error) {
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

func connectToDeploy(client *clientset.Clientset, namespace string) v1.DeploymentInterface {
	kubeclient := client.AppsV1().Deployments(namespace)
	return kubeclient
}

func createPod(config appsv1.Deployment, kubeclient v1.DeploymentInterface) error {
	// Manage resource
	_, err := kubeclient.Create(context.TODO(), &config, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Deployment Created successfully!")
	return nil
}

func deletePod(kubeclient v1.DeploymentInterface, name string) error {
	// Manage resource
	err := kubeclient.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Deployment Deleted successfully!")
	return nil
}

// Run some specific process, and it run as go routine.
func execCommand(config *rest.Config, clientset *clientset.Clientset, command string, pod corev1.Pod, stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := clientset.CoreV1().RESTClient().Post().Namespace(pod.Namespace).
		Name(pod.Name).Resource("pods").SubResource("exec")
	var option = &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(option, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Println(err)
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		log.Println(err)
	}
}
