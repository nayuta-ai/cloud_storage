package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func StressTest() {
	// test case
	var MEMORY_DEFAULT_LIMIT = 31457280
	var MEMORY_VPA_LIMIT = 104857600

	// Connect to Kubernetes
	config, client, err := connectToKubernetes()
	if err != nil {
		log.Printf("unexpected error: %v", err)
	}
	kubeclient := connectToDeploy(client, "default")

	// Create Pods
	err = createPod(*testObject, kubeclient)
	if err != nil {
		log.Printf("unexpected error: %v", err)
	}

	time.Sleep(60 * time.Second)

	// Fetch pod lists for getting pods name or container name
	pods, err := fetchPodList(client)
	if err != nil {
		log.Printf("unexpected error: %v", err)
	}

	// Execute stress command
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr
	go execCommand(config, client, "stress -m 1 --vm-bytes 52428800 --vm-hang 0", pods[0], stdin, stdout, stderr)
	time.Sleep(60 * time.Second)

	if !check_vpa(pods[0].Spec.Containers[0]) {
		log.Println("unexpected error: VPA doesn't work")
	}
	// Fetch container metrics after placing a load
	_, memory, err := fetchMetrics(config, pods[0].Spec.Containers[0].Name)
	if err != nil {
		log.Printf("unexpected error: %v", err)
	}
	fmt.Println(memory[1])
	if memory[1] > MEMORY_DEFAULT_LIMIT {
		log.Println("unexpected error: The memory resource is out of limits.")
	}
	fmt.Println(memory[0])
	if MEMORY_DEFAULT_LIMIT > memory[0] || memory[0] > MEMORY_VPA_LIMIT {
		log.Println("unexpected error: The memory resources is out of range.")
	}
	time.Sleep(10 * time.Second)
	err = deletePod(kubeclient, "sample-vpa-deployment")
	if err != nil {
		log.Printf("unexpected error: %v", err)
	}
}
