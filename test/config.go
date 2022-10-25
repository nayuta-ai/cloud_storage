package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
