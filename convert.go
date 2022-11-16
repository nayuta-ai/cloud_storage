package main

import (
	"io/ioutil"
	"os"

	v1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"
)

func createJobStructure(path string) (*v1.Job, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	job := &v1.Job{}
	err = yaml.Unmarshal(b, job)
	if err != nil {
		panic(err)
	}

	return job, nil
}
