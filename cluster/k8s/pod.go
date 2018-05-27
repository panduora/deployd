package k8s

import (
	"strings"

	"github.com/laincloud/deployd/model"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	operator "k8s.io/client-go/kubernetes/typed/core/v1"
)

type K8sPodCtrl struct {
	client operator.PodInterface
}

func NewK8sPod(cluster *K8sCluster, namespace string) *K8sPodCtrl {
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	return &K8sPodCtrl{
		client: cluster.Clientset.CoreV1().Pods(namespace),
	}
}

func (p *K8sPodCtrl) Inspect(pgs model.PodGroupSpec) *apiv1.PodList {
	podList, _ := p.client.List(metav1.ListOptions{LabelSelector: "app=console,deployer=LAIN"})
	return podList
}

func (p *K8sPodCtrl) Render(pgs model.PodGroupSpec) apiv1.PodTemplateSpec {
	return p.RenderPodTemplate(pgs)
}

func (p *K8sPodCtrl) RenderPodTemplate(pgs model.PodGroupSpec) apiv1.PodTemplateSpec {
	return apiv1.PodTemplateSpec{
		ObjectMeta: p.RenderPodMetaData(pgs.Pod),
		Spec:       p.RenderPodSpec(pgs.Pod),
	}
}

func (p *K8sPodCtrl) RenderPodMetaData(ps model.PodSpec) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: map[string]string{
			"app":      ps.Namespace,
			"proc":     strings.Replace(ps.Name, ".", "-", -1),
			"deployer": "LAIN",
		},
	}
}

func (p *K8sPodCtrl) RenderPodSpec(ps model.PodSpec) apiv1.PodSpec {
	return apiv1.PodSpec{
		Containers: p.RenderPodContainers(ps),
		Volumes:    p.RenderPodVolumes(ps),
	}
}

func (p *K8sPodCtrl) RenderPodContainers(ps model.PodSpec) []apiv1.Container {
	var containers []apiv1.Container

	for _, c := range ps.Containers {
		containers = append(containers, apiv1.Container{
			Name:         strings.Replace(ps.Name, ".", "-", -1),
			Image:        c.Image,
			Ports:        p.RenderContainerPort(c),
			Command:      c.Command,
			VolumeMounts: []apiv1.VolumeMount{},
		})
	}
	return containers
}

func (p *K8sPodCtrl) RenderContainerPort(cs model.ContainerSpec) []apiv1.ContainerPort {
	if cs.Expose > 0 {
		return []apiv1.ContainerPort{
			{
				Name:          "http",
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: int32(cs.Expose),
			},
		}
	} else {
		return []apiv1.ContainerPort{}
	}
}

func (p *K8sPodCtrl) RenderPodVolumes(ps model.PodSpec) []apiv1.Volume {
	return []apiv1.Volume{}
}
