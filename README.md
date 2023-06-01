# cloud_storage

start minikube
```
start minikube
```

enable metrics server addons
```
minikube addons enable metrics-server
```

get the metrics server
```
kubectl get deployment metrics-server -n kube-system
```

install the VPA
```
git clone https://github.com/kubernetes/autoscaler.git
cd autoscaler
git checkout vpa-release-0.8
./vertical-pod-autoscaler/hack/vpa-up.sh
```
