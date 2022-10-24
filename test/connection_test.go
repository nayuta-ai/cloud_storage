package main

import (
	"reflect"
	"testing"
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

func TestFetchPodLogs(t *testing.T) {
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
		log, err := fetchPodLogs(clientset, pod, containerList[0].Name)
		if err != nil {
			t.Fatalf("Failed for fetching pod log.")
		}
		testcase_string := reflect.TypeOf(log).Kind()
		if testcase_string != reflect.String {
			t.Errorf("The type of pod's log should be string.")
		}
	}
}
