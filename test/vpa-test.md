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
```
$ go run ./test
```