package k8s

import (
	"path/filepath"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/model"
	"github.com/mijia/adoc"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	K8sDefaultNamespace = "lain"
)

type K8sCluster struct {
	*kubernetes.Clientset

	*adoc.DockerClient
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
	CreateDeployment(c, spec)
	return model.PodGroup{}, nil
}

func (c *K8sCluster) RemovePodGroup(spec model.PodGroupSpec) error {
	return RemoveDeployment(c, spec)
}

func (c *K8sCluster) InspectPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func (c *K8sCluster) PatchPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func NewCluster(addr string, timeout, rwTimeout time.Duration, debug ...bool) (cluster.Cluster, error) {
	// FIXME: get dir from os env
	kubeconfig := filepath.Join("/root", ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	k8s := &K8sCluster{}
	k8s.Clientset = clientset
	k8s.DockerClient = &adoc.DockerClient{}
	return k8s, nil
}
