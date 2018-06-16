package k8s

import (
	"fmt"
	"strings"

	"github.com/laincloud/deployd/model"

	apiv1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
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
	labelSelector := fmt.Sprintf(
		"deployer=LAIN,app=%s,proc=%s", pgs.Namespace, strings.Replace(
			pgs.Name, ".", "-", -1))
	podList, _ := p.client.List(metav1.ListOptions{
		LabelSelector: labelSelector})
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
	// TODO: add the version info
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
			Env:          p.RenderContainerEnv(c),
			Resources:    p.RenderContainerResouces(c),
			VolumeMounts: []apiv1.VolumeMount{},
		})
	}
	return containers
}

func (p *K8sPodCtrl) RenderContainerPort(cs model.ContainerSpec) []apiv1.ContainerPort {
	if cs.Expose > 0 {
		return []apiv1.ContainerPort{
			{
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: int32(cs.Expose),
			},
		}
	} else {
		return []apiv1.ContainerPort{}
	}
}

func (p *K8sPodCtrl) RenderContainerEnv(cs model.ContainerSpec) []apiv1.EnvVar {

	envs := make([]apiv1.EnvVar, len(cs.Env))

	for i := range envs {
		kv := strings.Split(cs.Env[i], "=")
		envs[i] = apiv1.EnvVar{
			Name:  kv[0],
			Value: kv[1],
		}
	}

	return envs
}

func (p *K8sPodCtrl) RenderContainerResouces(cs model.ContainerSpec) apiv1.ResourceRequirements {
	// TODO: support cpu limits
	return apiv1.ResourceRequirements{
		Limits: apiv1.ResourceList{
			"memory": *resource.NewQuantity(
				cs.MemoryLimit, resource.BinarySI,
			),
		},
		Requests: apiv1.ResourceList{
			"memory": *resource.NewQuantity(
				cs.MemoryLimit, resource.BinarySI,
			),
		},
	}
}
func (p *K8sPodCtrl) RenderPodVolumes(ps model.PodSpec) []apiv1.Volume {
	return []apiv1.Volume{}
}
