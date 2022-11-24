package main

import (
	"cloud/vpatest"
	"log"
)

func ExampleStressTest() {
	var (
		vpapath  = "/home/yuta/cloud_storage/example/sample-vpa.yaml"
		filepath = "/home/yuta/cloud_storage/example/sample-deployment.yaml"
	)
	// Connect to Kubernetes
	cluster, err := vpatest.NewCluster()
	if err != nil {
		log.Printf("failed to create new cluster: %v", err)
	}
	err = cluster.NewMetricsClient()
	if err != nil {
		log.Printf("failed to create new client: %v", err)
	}
	err = cluster.NewClient()
	if err != nil {
		log.Printf("failed to create new client: %v", err)
	}
	// Create Pods
	config, pods, err := cluster.CreatePod(filepath, "")
	if err != nil {
		log.Fatalf("failed to create new pods: %v", err)
	}

	// Run VPA test
	err = cluster.Run(vpapath, pods[0])
	if err != nil {
		log.Println(err)
	}
	// Delete all pods for the test
	err = cluster.DeletePod("", config.ObjectMeta.Name)
	if err != nil {
		log.Printf("failed to delete the pods: %v", err)
	}
}
