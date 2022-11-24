package vpatest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/yaml"
)

func (cluster *Cluster) Run(vpapath string, pods corev1.Pod) error {
	vpaConfig, err := fetchVPAConfig(vpapath)
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
	go cluster.execCommand(stressCommand, pods)
	log.Println("load memory on pods")
	var memoryLoadVal int64
	for i := 0; ; i++ {
		_, memory, err := cluster.fetchMetricsList(pods.Spec.Containers[0].Name)
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
		fmt.Println(memoryLoadVal)
		fmt.Println(MEMORY_DEFAULT_LIMIT)
		fmt.Println(MEMORY_VPA_LIMIT)
		return errors.New("unexpected error: the memory resources is out of range")
	}
	return nil
}

func (cluster *Cluster) execCommand(command string, podList corev1.Pod) {
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

func (cluster *Cluster) CreatePod(filepath string, namespace string) (*appsv1.Deployment, []corev1.Pod, error) {
	if namespace == "" {
		namespace = "default"
	}
	deploymentInterface := cluster.newDeploymentInterface(namespace)
	if deploymentInterface == nil {
		return nil, nil, errors.New("nil pointer")
	}
	deploymentConfig, err := createDeploymentStructure(filepath)
	if err != nil {
		return nil, nil, err
	}
	_, err = deploymentInterface.Create(context.TODO(), deploymentConfig, metav1.CreateOptions{})
	if err != nil {
		return nil, nil, err
	}
	log.Println("Deployment Created successfully!")
	var pods []corev1.Pod
	for i := 0; ; i++ {
		if i > 10000 {
			return nil, nil, fmt.Errorf("time limit exceed: couldn't fetch the pod information")
		}
		tmpPods, err := cluster.fetchPodList("")
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

func (cluster *Cluster) newDeploymentInterface(namespace string) v1.DeploymentInterface {
	if cluster.Client.Clientset == nil {
		return nil
	}
	deployment := cluster.Client.Clientset.AppsV1().Deployments(namespace)
	return deployment
}

func createDeploymentStructure(path string) (*appsv1.Deployment, error) {
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

func (cluster *Cluster) DeletePod(namespace string, deploymentName string) error {
	if namespace == "" {
		namespace = "default"
	}
	deploymentInterface := cluster.newDeploymentInterface(namespace)
	if deploymentInterface == nil {
		return errors.New("nil pointer")
	}
	// Manage resource
	err := deploymentInterface.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Println("Deployment Deleted successfully!")
	return nil
}
