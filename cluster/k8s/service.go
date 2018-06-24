package k8s

import (
	"strings"

	"github.com/laincloud/deployd/model"
	"github.com/mijia/sweb/log"

	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	operator "k8s.io/client-go/kubernetes/typed/core/v1"
)

type K8sServiceCtrl struct {
	client  operator.ServiceInterface
	service *apiv1.Service
}

func NewK8sService(cluster *K8sCluster, namespace string) *K8sServiceCtrl {
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	return &K8sServiceCtrl{
		client: cluster.Clientset.CoreV1().Services(namespace),
	}
}

func (d *K8sServiceCtrl) Render(pgs model.PodGroupSpec) error {
	var ports []apiv1.ServicePort
	for _, c := range pgs.Pod.Containers {
		if c.Expose == 0 {
			continue
		}
		ports = append(ports, apiv1.ServicePort{
			Protocol:   apiv1.ProtocolTCP,
			Port:       int32(c.Expose),
			TargetPort: intstr.FromInt(c.Expose),
		})
	}

	d.service = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Replace(pgs.Name, ".", "-", -1),
		},
		Spec: apiv1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app":  pgs.Namespace,
				"proc": strings.Replace(pgs.Name, ".", "-", -1),
			},
		},
	}

	return nil
}

func (d *K8sServiceCtrl) Create(pgs model.PodGroupSpec) error {
	log.Infof("Creating Service...%q", pgs)
	d.Render(pgs)
	result, err := d.client.Create(d.service)
	log.Infof("Created Service %q.\n", result.GetObjectMeta().GetName())
	return err
}

func (d *K8sServiceCtrl) Upgrade(pgs model.PodGroupSpec) error {
	log.Infof("Upgrading Service...%q", pgs)
	name := strings.Replace(pgs.Name, ".", "-", -1)
	result, _ := d.client.Get(name, metav1.GetOptions{})
	log.Infof("Origin Service %s.\n", result)
	d.Render(pgs)
	result.Spec = d.service.Spec
	log.Infof("Patched Service %s.\n", result)
	_, updateErr := d.client.Update(result)
	log.Infof("Upgrading Service %q.\n", updateErr)
	return updateErr
}

func (d *K8sServiceCtrl) Remove(pgs model.PodGroupSpec) error {
	// FIXME: need scale down replica then remove Service
	log.Infof("Remving Service...%q", pgs)
	deletePolicy := metav1.DeletePropagationForeground
	pgName := strings.Replace(pgs.Name, ".", "-", -1)
	return d.client.Delete(pgName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}
