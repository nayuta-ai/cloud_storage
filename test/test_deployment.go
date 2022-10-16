package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inf.v0"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	kubeclient, err := connect_k8s("default")
	if err != nil {
		panic(err)
	}
	err = create_pod(*object, kubeclient)
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
	err = delete_pod(kubeclient)
	if err != nil {
		panic(err)
	}
}

func connect_k8s(namespacce string) (v1.DeploymentInterface, error) {
	// Create client
	var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kubeclient := client.AppsV1().Deployments(namespacce)
	return kubeclient, nil
}

var object = &appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-deployment",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: ptrint32(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "demo",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "demo",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "web",
						Image: "nginx:1.12",
						Ports: []corev1.ContainerPort{
							corev1.ContainerPort{
								Name:          "http",
								HostPort:      0,
								ContainerPort: 80,
								Protocol:      corev1.Protocol("TCP"),
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    *resource.NewDecimalQuantity(create_dec("300m"), resource.DecimalSI),
								"memory": *resource.NewQuantity(314572800, resource.BinarySI),
							},
							Requests: corev1.ResourceList{
								"cpu":    *resource.NewDecimalQuantity(create_dec("100m"), resource.DecimalSI),
								"memory": *resource.NewQuantity(104857600, resource.BinarySI),
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

func ptrint32(p int32) *int32 {
	return &p
}

func create_dec(cpu string) inf.Dec {
	base := inf.NewDec(1, 0)
	if strings.HasSuffix(cpu, "m") {
		base = inf.NewDec(1000, 0)
		cpu = strings.TrimRight(cpu, "m")
	}
	convertedStrUint64, _ := strconv.ParseUint(cpu, 10, 64)
	num_inf := inf.NewDec(int64(convertedStrUint64), 0)
	return *new(inf.Dec).QuoExact(num_inf, base)
}

func create_pod(config appsv1.Deployment, kubeclient v1.DeploymentInterface) error {
	// Manage resource
	_, err := kubeclient.Create(context.TODO(), &config, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Deployment Created successfully!")
	return nil
}

func delete_pod(kubeclient v1.DeploymentInterface) error {
	// Manage resource
	err := kubeclient.Delete(context.TODO(), "test-deployment", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Deployment Deleted successfully!")
	return nil
}
