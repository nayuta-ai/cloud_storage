# cloud_storage

start minikube
```
start minikube
```

get the metrics server
```
kubectl get deployment metrics-server -n kube-system
```

install the VPA
```
git clone https://github.com/kubernetes/autoscaler.git
./autoscaler/vertical-pod-autoscaler/hack/vpa-up.sh
```
