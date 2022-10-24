package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func connectToDeploy(client *clientset.Clientset, namespace string) v1.DeploymentInterface {
	kubeclient := client.AppsV1().Deployments(namespace)
	return kubeclient
}

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

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, client, nil
}

var testObject = &appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "sample-vpa-deployment",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrint32(2),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "sample-app",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "sample-app",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "vpa-container",
						Image: "amsy810/tools:v2.0",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    *resource.NewDecimalQuantity(convertDec("30m"), resource.DecimalSI),
								"memory": *resource.NewQuantity(31457280, resource.BinarySI),
							},
							Requests: corev1.ResourceList{
								"cpu":    *resource.NewDecimalQuantity(convertDec("10m"), resource.DecimalSI),
								"memory": *resource.NewQuantity(10485760, resource.BinarySI),
							},
						},
					},
				},
			},
		},
		Strategy:        appsv1.DeploymentStrategy{},
		MinReadySeconds: 0,
	},
}

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

func fetchPodList(clientset *clientset.Clientset) ([]corev1.Pod, error) {
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func fetchContainerlist(pod corev1.Pod) []corev1.Container {
	var containerList []corev1.Container
	containerList = append(containerList, pod.Spec.Containers...)
	return containerList
}

func fetchPodLogs(clientset *clientset.Clientset, pod corev1.Pod, container string) (string, error) {
	podLopOpts := corev1.PodLogOptions{}
	podLopOpts.Container = container
	podLopOpts.TailLines = &[]int64{int64(100)}[0]
	podLopOpts.Follow = true
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLopOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	log := buf.String()
	fmt.Println(log)
	return log, nil
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

func execCommand(config *rest.Config, clientset *clientset.Clientset, command string, pod corev1.Pod, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
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
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}
	return nil
}
func check_vpa(container corev1.Container) bool {
	return fmt.Sprintf("%v", container.Resources.Requests.Memory()) == fmt.Sprintf("%v", resource.NewQuantity(62914560, resource.BinarySI))
}
