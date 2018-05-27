package k8s

import (
	"strings"

	"github.com/laincloud/deployd/model"
	"github.com/mijia/sweb/log"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	operator "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
)

type K8sDeploymentCtrl struct {
	client     operator.DeploymentInterface
	deployment *appsv1beta1.Deployment

	podCtrl *K8sPodCtrl
}

func NewK8sDeployment(cluster *K8sCluster, namespace string) *K8sDeploymentCtrl {
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	return &K8sDeploymentCtrl{
		client:  cluster.Clientset.AppsV1beta1().Deployments(namespace),
		podCtrl: NewK8sPod(cluster, namespace),
	}
}

func (d *K8sDeploymentCtrl) Render(pgs model.PodGroupSpec) error {
	d.deployment = &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Replace(pgs.Name, ".", "-", -1),
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(int32(pgs.NumInstances)),
			Template: d.podCtrl.Render(pgs),
		},
	}

	return nil
}

func (d *K8sDeploymentCtrl) Create(pgs model.PodGroupSpec) error {
	log.Infof("Creating deployment...%q", pgs)
	d.Render(pgs)
	result, err := d.client.Create(d.deployment)
	log.Infof("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return err
}

func (d *K8sDeploymentCtrl) Remove(pgs model.PodGroupSpec) error {
	// FIXME: need scale down replica then remove deployment
	log.Infof("Remving deployment...%q", pgs)
	deletePolicy := metav1.DeletePropagationForeground
	pgName := strings.Replace(pgs.Name, ".", "-", -1)
	return d.client.Delete(pgName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func (d *K8sDeploymentCtrl) Inspect(pgs model.PodGroupSpec) error {
	// FIXME: need scale down replica then remove deployment
	log.Infof("Inspecting deployment...%q", pgs)

	podList := d.podCtrl.Inspect(pgs)
	for _, p := range podList.Items {
		log.Infof("Pod %q", p.Name)
		log.Infof("Pod host IP of status %q", p.Status.HostIP)
		log.Infof("Pod status %q", p.Status)
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
