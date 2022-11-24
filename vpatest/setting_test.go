package vpatest_test

import (
	"cloud/vpatest"
	"testing"
)

var cluster *vpatest.Cluster

func init() {
	testing.Init()
	c, err := vpatest.NewCluster()
	if err != nil {
		panic(err)
	}
	cluster = c
}
