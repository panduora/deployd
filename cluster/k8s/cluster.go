package k8s

import (
	"fmt"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/model"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	K8sDefaultNamespace = "lain"
)

type K8sCluster struct {
	*kubernetes.Clientset
}

// GetResources get node info
func (c *K8sCluster) GetResources() ([]cluster.Node, error) {
	return nil, nil
}

func (c *K8sCluster) ListPodGroups(showAll bool, filters ...string) ([]model.PodGroup, error) {
	return nil, nil
}

func (c *K8sCluster) CreatePodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	// 1. init podctls, podgroup
	// 2. use podctl deploy instance
	// 3. assemble podgroup
	return model.PodGroup{}, nil
}

func (c *K8sCluster) InspectPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func (c *K8sCluster) PatchPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func NewCluster(addr string, timeout, rwTimeout time.Duration, debug ...bool) (cluster.Cluster, error) {
	home := homeDir()
	kubeconfig = filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	k8s := &K8sCluster{}
	k8s.Clientset = clientset
	return k8s, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
}
