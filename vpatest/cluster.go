package vpatest

import (
	"flag"
	"os"
	"path/filepath"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Cluster struct {
	Config *rest.Config
	Client Client
}

type Client struct {
	Metrics   *metrics.Clientset
	Clientset *clientset.Clientset
}

// NewCluster connects a cluster.
func NewCluster() (*Cluster, error) {
	// The first step is to connect to the cluster.
	// Connection to the cluster requires K8s cluster config file, since we are connecting to an external remote cluster.
	// Since it is assumed that you have a cluster set-up, you must be having the config file ready.
	// It is assumed that the cluster config file is located at $HOME/.kube/config
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// The next step is to build the cluster config from url or config file path.
	// In this case, it builds the cluster config from config file path.
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return &Cluster{
		Config: config,
	}, nil
}

// NewMetricsClient creates *metrics.Clientset from *rest.Config.
// It can get the resource information on the pods.
func (cluster *Cluster) NewMetricsClient() error {
	mc, err := metrics.NewForConfig(cluster.Config)
	if err != nil {
		return err
	}
	cluster.Client.Metrics = mc
	return nil
}

// NewClient creates the clientset from *rest.Config.
// It can get the some information from pods or nodes.
func (cluster *Cluster) NewClient() error {
	client, err := clientset.NewForConfig(cluster.Config)
	if err != nil {
		return err
	}
	cluster.Client.Clientset = client
	return nil
}

// homeDir gets the home directory.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
