package k8s

import (
	"strings"

	"github.com/laincloud/deployd/model"
	"github.com/mijia/sweb/log"

	apps "k8s.io/api/apps/v1beta2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	operator "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
)

type K8sStatefulSetCtrl struct {
	client      operator.StatefulSetInterface
	statefulSet *apps.StatefulSet

	podCtrl *K8sPodCtrl
}

func NewK8sStatefulSet(cluster *K8sCluster, namespace string) *K8sStatefulSetCtrl {
	if namespace == "" {
		namespace = core.NamespaceDefault
	}
	return &K8sStatefulSetCtrl{
		client:  cluster.Clientset.AppsV1beta2().StatefulSets(namespace),
		podCtrl: NewK8sPod(cluster, namespace),
	}
}

func (d *K8sStatefulSetCtrl) Render(pgs model.PodGroupSpec) error {
	d.statefulSet = &apps.StatefulSet{
		ObjectMeta: meta.ObjectMeta{
			Name: strings.Replace(pgs.Name, ".", "-", -1),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:             int32Ptr(int32(pgs.NumInstances)),
			Template:             d.podCtrl.Render(pgs),
			VolumeClaimTemplates: nil,
		},
	}

	return nil
}

func (d *K8sStatefulSetCtrl) Create(pgs model.PodGroupSpec) error {
	log.Infof("Creating Statefulset...%q", pgs)
	d.Render(pgs)
	result, err := d.client.Create(d.statefulSet)
	log.Infof("Created Statefulset %q.\n", result.GetObjectMeta().GetName())
	return err
}

func (d *K8sStatefulSetCtrl) Remove(pgs model.PodGroupSpec) error {
	// FIXME: need scale down replica then remove Statefulset
	log.Infof("Remving Statefulset...%q", pgs)
	deletePolicy := meta.DeletePropagationForeground
	pgName := strings.Replace(pgs.Name, ".", "-", -1)
	return d.client.Delete(pgName, &meta.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func (d *K8sStatefulSetCtrl) Inspect(pgs model.PodGroupSpec) model.PodGroup {
	// FIXME: need scale down replica then remove Statefulset
	log.Infof("Inspecting Statefulset...%q", pgs)

	podList := d.podCtrl.Inspect(pgs)
	for _, p := range podList.Items {
		log.Infof("Pod %q", p.Name)
		log.Infof("Pod host IP of status %q", p.Status.HostIP)
		log.Infof("Pod status %q", p.Status)
	}

	return model.PodGroup{}
}
