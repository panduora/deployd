package k8s

import (
	"github.com/laincloud/deployd/model"
)

type K8sWorkloadCtrl interface {
	Render(pgs model.PodGroupSpec) error
	Create(pgs model.PodGroupSpec) error
	Remove(pgs model.PodGroupSpec) error
	Inspect(pgs model.PodGroupSpec) model.PodGroup
}

func NewWorkload(cluster *K8sCluster, spec model.PodGroupSpec, namespace string) K8sWorkloadCtrl {
	if spec.Pod.IsStateful() {
		return NewK8sStatefulSet(cluster, namespace)
	} else {
		return NewK8sDeployment(cluster, namespace)
	}
}
