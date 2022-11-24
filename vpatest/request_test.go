package vpatest_test

import (
	"testing"
)

type Result struct {
	ObjectMeta    string
	ContainerName string
}

func TestNewCreatePod(t *testing.T) {
	err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	var test = []struct {
		name     string
		filepath string
		want     Result
	}{
		{
			name:     "test1",
			filepath: "/home/yuta/cloud_storage/example/sample-deployment.yaml",
			want: Result{
				"sample-vpa-deployment",
				"vpa-container",
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			deploymentStructure, pods, err := cluster.CreatePod(tt.filepath, "")
			if err != nil {
				t.Fatal(err)
			}
			if len(pods) == 0 {
				t.Error("failed to fetch pods")
			}
			if deploymentStructure.ObjectMeta.Name != tt.want.ObjectMeta {
				t.Errorf("The meta data name should be %s", tt.want.ObjectMeta)
			}
			if deploymentStructure.Spec.Template.Spec.Containers[0].Name != tt.want.ContainerName {
				t.Errorf("The container name should be %s", tt.want.ContainerName)
			}
			err = cluster.DeletePod("", deploymentStructure.ObjectMeta.Name)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
