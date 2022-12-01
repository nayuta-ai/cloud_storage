package vpatest_test

import (
	"cloud/vpatest"
	"testing"
)

var cluster *vpatest.Cluster

func init() {
	testing.Init()
	c, err := newTestCluster()
	if err != nil {
		panic(err)
	}
	cluster = c
}

func newTestCluster() (*vpatest.Cluster, error) {
	return &vpatest.Cluster{
		Model: &ConfigTest{},
	}, nil
}
