package vpatest_test

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	err := cluster.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if cluster.Client.Clientset == nil {
		t.Fatal("failed to create clientset")
	}
}

func TestNewMetricsClient(t *testing.T) {
	err := cluster.NewMetricsClient()
	if err != nil {
		t.Fatal(err)
	}
	if cluster.Client.Metrics == nil {
		t.Fatal("failed to create metrics client")
	}
}
