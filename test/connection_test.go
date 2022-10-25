package main

import (
	"reflect"
	"testing"

	"gopkg.in/inf.v0"
)

func TestFetchPodList(t *testing.T) {
	_, clientset, err := connectToKubernetes()
	if err != nil {
		t.Fatalf("Failed for connecting cluster.")
	}
	podsList, err := fetchPodList(clientset)
	if err != nil {
		t.Fatalf("Failed for fetching pod list.")
	}
	testcase_slice := reflect.TypeOf(podsList).Kind()
	if testcase_slice != reflect.Slice {
		t.Errorf("The type of podsList should be slice.")
	}
	for _, pod := range podsList {
		testcase_string := reflect.TypeOf(pod.Name).Kind()
		if testcase_string != reflect.String {
			t.Errorf("The type of pod's name should be string.")
		}
	}
}

func TestFetchMatrics(t *testing.T) {
	UPPER_LIMIT_CPU := inf.NewDec(100, 0)
	BASE_N := inf.NewDec(1000000000, 0) // 10**9
	UPPER_LIMIT_MEMORY := 62914560
	config, _, err := connectToKubernetes()
	if err != nil {
		t.Fatalf("Failed for connecting cluster.")
	}
	CONTAINER_NAME := "vpa-container"
	cpu, memory, err := fetchMetrics(config, CONTAINER_NAME)
	if err != nil {
		t.Errorf("%s", err)
	}
	for i := 0; i < len(cpu); i++ {
		upper := new(inf.Dec).QuoExact(UPPER_LIMIT_CPU, BASE_N)
		// upper < cpu[i]
		if upper.Cmp(cpu[i]) != -1 {
			t.Errorf("CPU resources is out of range.:%f", cpu[i])
		}
		if memory[i] > UPPER_LIMIT_MEMORY {
			t.Errorf("Memory resources is out of range.:%d", memory[i])
		}
	}
}

func TestVPA(t *testing.T) {
	_, clientset, err := connectToKubernetes()
	if err != nil {
		t.Fatalf("Failed for connecting cluster.")
	}
	podsList, err := fetchPodList(clientset)
	if err != nil {
		t.Fatalf("Failed for fetching pod list.")
	}
	for _, pod := range podsList {
		containerList := fetchContainerlist(pod)
		for _, container := range containerList {
			if !check_vpa(container) {
				t.Errorf("Failed for VPA")
			}
		}
	}
}
