package k8s

import (
	"fmt"
	"github.com/laincloud/deployd/model"
	"github.com/mijia/sweb/log"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDeployment(cluster *K8sCluster, pgs model.PodGroupSpec) {
	// Create Deployment
	fmt.Println("Creating deployment...")
	log.Infof("Creating deployment...%s ", pgs)
	deploymentsClient := cluster.Clientset.AppsV1beta1().Deployments(apiv1.NamespaceDefault)
	deployment := RenderDeployment(pgs)
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	log.Infof("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func RenderDeployment(pgs model.PodGroupSpec) *appsv1beta1.Deployment {
	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: pgs.Name,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(int32(pgs.NumInstances)),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: RenderPodMetaData(pgs.Pod),
				Spec:       RenderPodSpec(pgs.Pod),
			},
		},
	}
}

func RenderPodMetaData(ps model.PodSpec) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: map[string]string{
			"app":      ps.Namespace,
			"proc":     ps.Name,
			"deployer": "LAIN",
		},
	}
}

func RenderPodSpec(ps model.PodSpec) apiv1.PodSpec {
	return apiv1.PodSpec{
		Containers: RenderPodContainers(ps),
		Volumes:    RenderPodVolumes(ps),
	}
}

func RenderPodContainers(ps model.PodSpec) []apiv1.Container {
	var containers []apiv1.Container

	for _, c := range ps.Containers {
		containers = append(containers, apiv1.Container{
			Name:         ps.Name,
			Image:        c.Image,
			Ports:        RenderContainerPort(c),
			VolumeMounts: []apiv1.VolumeMount{},
		})
	}
	return containers
}

func RenderContainerPort(cs model.ContainerSpec) []apiv1.ContainerPort {
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

func RenderPodVolumes(ps model.PodSpec) []apiv1.Volume {
	return []apiv1.Volume{}
}

func int32Ptr(i int32) *int32 { return &i }
