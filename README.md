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
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

install the VPA
```
git clone https://github.com/kubernetes/autoscaler.git
cd autoscaler
git checkout vpa-release-0.8
./vertical-pod-autoscaler/hack/vpa-up.sh
```
