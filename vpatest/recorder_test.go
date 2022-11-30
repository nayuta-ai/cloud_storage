package vpatest_test

import (
	"cloud/vpatest"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/inf.v0"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalev1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// Config plays the role of setting file for functions in recorder.go
type ConfigTest struct{}

func (c *ConfigTest) FetchPodList(namespace string, cluster *vpatest.Cluster) ([]corev1.Pod, error) {
	podList := []corev1.Pod{
		{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "container1",
						Image: "TestImage",
					},
					{
						Name:  "container2",
						Image: "TestImage",
					},
				},
			},
		},
	}
	return podList, nil
}

func (c *ConfigTest) FetchMetricsList(containerName string, cluster *vpatest.Cluster) ([]*inf.Dec, []int64, error) {
	cpuList := []*inf.Dec{newDec("30m")}
	memoryList := []int64{54857600}
	return cpuList, memoryList, nil
}

func newDec(cpu string) *inf.Dec {
	base := inf.NewDec(1, 0)
	if strings.HasSuffix(cpu, "m") {
		base = inf.NewDec(1000, 0)
		cpu = strings.TrimRight(cpu, "m")
	}
	convertedStrUint64, _ := strconv.ParseUint(cpu, 10, 64)
	num_inf := inf.NewDec(int64(convertedStrUint64), 0)
	return new(inf.Dec).QuoExact(num_inf, base)
}

func TestCreateDeploymentStructure(t *testing.T) {
	pod := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-vpa-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "sample-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
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
		},
	}
	var test = []struct {
		name     string
		filepath string
		want     *appsv1.Deployment
	}{
		{
			name:     "test1",
			filepath: "/home/yuta/cloud_storage/example/sample-deployment.yaml",
			want:     pod,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			pods, err := vpatest.FetchDeploymentConfig(tt.filepath)
			if err != nil {
				t.Fatal(err)
			}
			if pods.ObjectMeta.Name != tt.want.ObjectMeta.Name {
				t.Errorf("the pods name should be %s", tt.want.ObjectMeta.Name)
			}
			if pods.Spec.Template.Spec.Containers[0].Name != tt.want.Spec.Template.Spec.Containers[0].Name {
				t.Errorf("the container name should be %s", tt.want.Spec.Template.Spec.Containers[0].Name)
			}
			if pods.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() != tt.want.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() {
				t.Errorf("the cpu resource should be %s", tt.want.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu())
			}
			if pods.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String() != tt.want.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String() {
				t.Errorf("the memory resource should be %s", tt.want.Spec.Template.Spec.Containers[0].Resources.Limits.Memory())
			}
			if pods.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String() != tt.want.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String() {
				t.Errorf("the cpu resource should be %s", tt.want.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu())
			}
			if pods.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String() != tt.want.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String() {
				t.Errorf("the memory resource should be %s", tt.want.Spec.Template.Spec.Containers[0].Resources.Requests.Memory())
			}
		})
	}
}

func TestFetchVPAConfig(t *testing.T) {
	var vpaConfig = &autoscalev1.VerticalPodAutoscaler{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: autoscalev1.VerticalPodAutoscalerSpec{
			TargetRef: &v1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "sample-vpa-deployment",
				APIVersion: "apps/v1",
			},
			UpdatePolicy: &autoscalev1.PodUpdatePolicy{},
			ResourcePolicy: &autoscalev1.PodResourcePolicy{
				ContainerPolicies: []autoscalev1.ContainerResourcePolicy{
					{
						ContainerName: "no-vpa-container",
					},
					{
						ContainerName: "*",
						MinAllowed: corev1.ResourceList{
							"cpu":    *resource.NewDecimalQuantity(convertDec("10m"), resource.DecimalSI),
							"memory": *resource.NewQuantity(31457280, resource.BinarySI),
						},
						MaxAllowed: corev1.ResourceList{
							"cpu":    *resource.NewDecimalQuantity(convertDec("100m"), resource.DecimalSI),
							"memory": *resource.NewQuantity(62914560, resource.BinarySI),
						},
					},
				},
			},
		},
	}
	var test = []struct {
		name string
		file string
		want *autoscalev1.VerticalPodAutoscaler
	}{
		{
			name: "test1",
			file: "/home/yuta/cloud_storage/example/sample-vpa.yaml",
			want: vpaConfig,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			vpa, err := vpatest.FetchVPAConfig(tt.file)
			if err != nil {
				t.Error(err)
			}
			if vpa.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Cpu() == tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Cpu() {
				t.Errorf("the min cpu resource should be %s", tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Cpu())
			}
			if vpa.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Memory() == tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Memory() {
				t.Errorf("the min memory resource should be %s", tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MinAllowed.Memory())
			}
			if vpa.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Cpu() == tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Cpu() {
				t.Errorf("the max cpu resource should be %s", tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Cpu())
			}
			if vpa.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Memory() == tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Memory() {
				t.Errorf("the max memory resource should be %s", tt.want.Spec.ResourcePolicy.ContainerPolicies[1].MaxAllowed.Memory())
			}
		})
	}
}
