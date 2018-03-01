package swarm

import (
	"fmt"
	"time"

	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/model"
	"github.com/mijia/adoc"
)

type SwarmCluster struct {
	*adoc.DockerClient
}

func (c *SwarmCluster) GetResources() ([]cluster.Node, error) {
	if info, err := c.DockerClient.SwarmInfo(); err != nil {
		return nil, err
	} else {
		nodes := make([]cluster.Node, len(info.Nodes))
		for i, node := range info.Nodes {
			nodes[i] = cluster.Node{
				Name:       node.Name,
				Address:    node.Address,
				Containers: node.Containers,
				CPUs:       node.CPUs,
				UsedCPUs:   node.UsedCPUs,
				Memory:     node.Memory,
				UsedMemory: node.UsedMemory,
			}
		}
		return nodes, nil
	}
}

func (c *SwarmCluster) ListPodGroups(showAll bool, filters ...string) ([]model.PodGroup, error) {
	return nil, nil
}

func (c *SwarmCluster) CreatePodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	// 1. init podctls, podgroup
	// 2. use podctl deploy instance
	// 3. assemble podgroup
	return model.PodGroup{}, nil
}

func (c *SwarmCluster) InspectPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func (c *SwarmCluster) PatchPodGroup(spec model.PodGroupSpec) (model.PodGroup, error) {
	return model.PodGroup{}, nil
}

func NewCluster(addr string, timeout, rwTimeout time.Duration, debug ...bool) (cluster.Cluster, error) {
	docker, err := adoc.NewSwarmClientTimeout(addr, nil, timeout, rwTimeout)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect swarm master[%s], %s", addr, err)
	}
	if len(debug) > 0 && debug[0] {
		adoc.EnableDebug()
	}
	swarm := &SwarmCluster{}
	swarm.DockerClient = docker
	return swarm, nil
}
