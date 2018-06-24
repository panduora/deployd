package k8s

import (
	"strings"
	"time"

	"github.com/laincloud/deployd/model"
	"github.com/mijia/adoc"
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
	svcCtrl *K8sServiceCtrl
}

func NewK8sDeployment(cluster *K8sCluster, namespace string) *K8sDeploymentCtrl {
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	return &K8sDeploymentCtrl{
		client:  cluster.Clientset.AppsV1beta1().Deployments(namespace),
		podCtrl: NewK8sPod(cluster, namespace),
		svcCtrl: NewK8sService(cluster, namespace),
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

func (d *K8sDeploymentCtrl) Create(pgs model.PodGroupSpec) model.PodGroup {
	log.Infof("Creating deployment...%q", pgs)
	d.Render(pgs)
	d.svcCtrl.Create(pgs)
	result, _ := d.client.Create(d.deployment)
	log.Infof("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return d.Watch(pgs)
}

func (d *K8sDeploymentCtrl) Upgrade(pgs model.PodGroupSpec) model.PodGroup {
	log.Infof("Upgrading deployment...%q", pgs)
	name := strings.Replace(pgs.Name, ".", "-", -1)
	result, _ := d.client.Get(name, metav1.GetOptions{})
	log.Infof("Origin deployment %s.\n", result)
	d.Render(pgs)
	result.Spec = d.deployment.Spec
	log.Infof("Patched deployment %s.\n", result)
	d.svcCtrl.Upgrade(pgs)
	_, updateErr := d.client.Update(result)
	log.Infof("Upgrading deployment %q.\n", updateErr)
	return d.Watch(pgs)
}

func (d *K8sDeploymentCtrl) Remove(pgs model.PodGroupSpec) error {
	// FIXME: need scale down replica then remove deployment
	log.Infof("Remving deployment...%q", pgs)
	deletePolicy := metav1.DeletePropagationForeground
	pgName := strings.Replace(pgs.Name, ".", "-", -1)
	d.svcCtrl.Remove(pgs)
	return d.client.Delete(pgName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func (d *K8sDeploymentCtrl) Inspect(pgs model.PodGroupSpec) model.PodGroup {
	log.Infof("Inspecting deployment...%q", pgs)
	kPodList := d.podCtrl.Inspect(pgs)

	pg := model.PodGroup{}
	pg.State = model.RunStateSuccess
	pg.LastError = ""
	pg.Pods = make([]model.Pod, pgs.NumInstances)

	for i, kPod := range kPodList.Items {
		if i >= pgs.NumInstances {
			break
		}
		pod := model.Pod{}
		pod.InstanceNo = i + 1
		containers := make([]model.Container, len(kPod.Status.ContainerStatuses))

		for j := range containers {
			container := model.Container{}
			container.ContainerIp = kPod.Status.PodIP
			container.NodeIp = kPod.Status.HostIP
			container.Id = kPod.Status.ContainerStatuses[j].ContainerID
			container.NodeName = kPod.Spec.NodeName
			// FIXME: replace adoc container info with k8s related info
			container.Runtime = adoc.ContainerDetail{
				Id:    kPod.Status.ContainerStatuses[j].ContainerID,
				Image: kPod.Spec.Containers[j].Image,
				State: adoc.ContainerState{
					Running:   true,
					StartedAt: kPod.Status.StartTime.Time,
				},
				Name: kPod.Name,
			}
			containers[j] = container
		}
		pod.Containers = containers

		pg.Pods[i] = pod
	}

	return pg
}

func (d *K8sDeploymentCtrl) Watch(pgs model.PodGroupSpec) model.PodGroup {
	// FIXME: need more accurate watch way
	for {
		time.Sleep(5 * time.Second)

		pg := d.Inspect(pgs)
		validPod := 0
		for _, pod := range pg.Pods {
			if len(pod.Containers) > 0 {
				validPod += 1
			}
		}

		if validPod != pgs.NumInstances {
			continue
		}
		return pg
	}
}

func int32Ptr(i int32) *int32 { return &i }
