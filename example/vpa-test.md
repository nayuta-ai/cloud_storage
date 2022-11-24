## How to Test
1. start minikube
```
$ minikube start
```

2. start vpa
```
$ kubectl apply -f test/sample-vpa.yaml
```
If metrics server doesn't run, set the vpa according to READNE.md

3. start test
You set up the yaml file such as sample-deployment.yaml and sample-vpa.yaml.
```
$ go run ./example
```