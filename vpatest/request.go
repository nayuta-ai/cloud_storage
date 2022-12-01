package vpatest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

// Run executes stress command, fetches the resource information in pods, and validate the correct range of the resource fetched.
func (cluster *Cluster) Run(vpapath string, pods corev1.Pod) error {
	vpaConfig, err := FetchVPAConfig(vpapath)
	if err != nil {
		log.Println(err)
	}
	// test case
	MEMORY_VPA_LIMIT, ok := vpaConfig.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Memory().AsInt64()
	if !ok {
		return errors.New("failed to convert memory resources into int64")
	}
	MEMORY_DEFAULT_LIMIT, ok := vpaConfig.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Memory().AsInt64()
	if !ok {
		return errors.New("failed to convert memory resources into int64")
	}
	log.Println("fetch the vpa config")
	loadVal := (MEMORY_DEFAULT_LIMIT + MEMORY_VPA_LIMIT) / 2
	stressCommand := fmt.Sprintf("stress -m 1 --vm-bytes %d --vm-hang 0", loadVal)
	log.Println("used the right command:", stressCommand)
	go cluster.ExecCommand(stressCommand, pods)
	log.Println("load memory on pods")
	var memoryLoadVal int64
	for i := 0; ; i++ {
		_, memory, err := cluster.Model.FetchMetricsList(pods.Spec.Containers[0].Name, cluster)
		if err != nil {
			return err
		}
		if len(memory) != 0 {
			memoryLoadVal = memory[0]
			break
		}
	}
	log.Println("Fetched the pod metrics successfully")
	// Test for pods which stress process affects
	if MEMORY_DEFAULT_LIMIT > memoryLoadVal || memoryLoadVal > MEMORY_VPA_LIMIT {
		return errors.New("unexpected error: the memory resources is out of range")
	}
	return nil
}

// ExecCommand executes some command to a container in the podList.
func (cluster *Cluster) ExecCommand(command string, podList corev1.Pod) {
	time.Sleep(5 * time.Second)
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := cluster.Client.Clientset.CoreV1().RESTClient().Post().Namespace(podList.Namespace).
		Name(podList.Name).Resource("pods").SubResource("exec")
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr
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
	exec, err := remotecommand.NewSPDYExecutor(cluster.Config, "POST", req.URL())
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

// CreateePods creates a pod.
func (cluster *Cluster) CreatePod(filepath string, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}
	deploymentInterface := cluster.Client.Clientset.AppsV1().Deployments(namespace)
	if deploymentInterface == nil {
		return nil, errors.New("nil pointer")
	}
	deploymentConfig, err := FetchDeploymentConfig(filepath)
	if err != nil {
		return nil, err
	}
	_, err = deploymentInterface.Create(context.TODO(), deploymentConfig, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Println("Deployment Created successfully!")
	return deploymentConfig, nil
}

// CompletePods returns anything when creating new pods sucessfuly.
func (cluster *Cluster) CompletePods(filepath string, namespace string) (*appsv1.Deployment, []corev1.Pod, error) {
	deploymentConfig, err := cluster.CreatePod(filepath, namespace)
	if err != nil {
		return nil, nil, err
	}
	var pods []corev1.Pod
	for i := 0; ; i++ {
		if i > 10000 {
			return nil, nil, fmt.Errorf("time limit exceed: couldn't fetch the pod information")
		}
		tmpPods, err := cluster.Model.FetchPodList("", cluster)
		if err != nil {
			return nil, nil, err
		}
		if len(tmpPods) != 0 {
			pods = tmpPods
			break
		}
	}
	log.Println("Pods Created successfully")
	return deploymentConfig, pods, nil
}

// DeletePod delete an targeted pod.
func (cluster *Cluster) DeletePod(namespace string, deploymentName string) error {
	if namespace == "" {
		namespace = "default"
	}
	deploymentInterface := cluster.Client.Clientset.AppsV1().Deployments(namespace)
	// Manage resource
	err := deploymentInterface.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Println("Deployment Deleted successfully!")
	return nil
}
