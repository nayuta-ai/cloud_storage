package vpatest_test

import (
	"strconv"
	"strings"
	"testing"

	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCompletePods(t *testing.T) {
	var test = []struct {
		name     string
		filepath string
	}{
		{
			name:     "test1",
			filepath: "/home/yuta/cloud_storage/example/sample-deployment.yaml",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			cluster.Client.Clientset = testclient.NewSimpleClientset()
			_, _, err := cluster.CompletePods(tt.filepath, "")
			if err != nil {
				t.Error("failed to create pods:", err)
			}
		})
	}
}

func TestCreatePod(t *testing.T) {
	var test = []struct {
		name     string
		filepath string
	}{
		{
			name:     "test1",
			filepath: "/home/yuta/cloud_storage/example/sample-deployment.yaml",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			cluster.Client.Clientset = testclient.NewSimpleClientset()
			_, err := cluster.CreatePod(tt.filepath, "")
			if err != nil {
				t.Error("failed to create pods:", err)
			}
		})
	}
}

func convertDec(cpu string) inf.Dec {
	base := inf.NewDec(1, 0)
	if strings.HasSuffix(cpu, "m") {
		base = inf.NewDec(1000, 0)
		cpu = strings.TrimRight(cpu, "m")
	}
	convertedStrUint64, _ := strconv.ParseUint(cpu, 10, 64)
	num_inf := inf.NewDec(int64(convertedStrUint64), 0)
	return *new(inf.Dec).QuoExact(num_inf, base)
}

func TestRun(t *testing.T) {
	var test = []struct {
		name string
		path string
	}{
		{
			name: "test1",
			path: "/home/yuta/cloud_storage/example/sample-vpa.yaml",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			err := cluster.Run(tt.path, corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "pod1",
						},
					},
				},
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDeletePods(t *testing.T) {
	var test = []struct {
		name           string
		deploymentName string
	}{
		{
			name:           "test1",
			deploymentName: "sample-vpa-deployment",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			err := cluster.DeletePod("", tt.deploymentName)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
