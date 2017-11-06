package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/laincloud/deployd/cluster"
	"github.com/laincloud/deployd/storage"
	"github.com/mijia/sweb/log"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
)

const (
	CalicoSockPath = "/var/run/docker/plugins/calico.sock"
	DockerSockPath = "/var/run/docker.sock"
)

type ErrorSpec struct {
	Network       string
	ContainerId   string
	ContainerName string
	PrevNode      string
	PrevIP        string
	Message       string
}

func (es ErrorSpec) String() string {
	return fmt.Sprintf("Error[Network=%s, ID=%s, Name=%s, PrevIP=%s, Message=%s]",
		es.Network, es.ContainerId, es.ContainerName, es.PrevIP, es.Message)
}

type ContainerNetwork struct {
	Name        string
	EndpointID  string
	MacAddress  string
	IPv4Address string
	IPv6Address string
}

func (cn ContainerNetwork) String() string {
	return fmt.Sprintf("ContainerNetwork[Name=%s, EndpointID=%s, IPv4Address=%s]",
		cn.Name, cn.EndpointID, cn.IPv4Address)
}

type DockerNetwork struct {
	Name       string
	Id         string
	Containers map[string]ContainerNetwork
}

func (dn DockerNetwork) String() string {
	return fmt.Sprintf("DockerNetwork[Name=%s, ID=%s, Containers=%v]",
		dn.Name, dn.Id, len(dn.Containers))
}

type CalicoHandler struct {
	DockerClient  http.Client
	ClusterClient cluster.Cluster
	CalicoClient  http.Client
	StorageClient storage.Store
}

func (ch *CalicoHandler) Handle(es ErrorSpec) {
	errMsg := es.Message
	log.Infof("receive error spec %s", es)
	go ch.HandleNoSandboxPresent(es, errMsg)
	go ch.HandleEndpointExist(es, errMsg)
	go ch.HandleIPInUse(es, errMsg)
}

func (ch *CalicoHandler) HandleNoSandboxPresent(es ErrorSpec, errMsg string) {
	if NoSandboxPresentError(errMsg) == false {
		return
	}

	network, containerId := es.Network, es.ContainerId
	err := ch.ClusterClient.DisconnectContainer(network, containerId, true)
	if err != nil {
		log.Errorf("disconnect container %s from network %s failed: %s",
			containerId, network, err)
	}

	err = ch.ClusterClient.ConnectContainer(network, containerId, es.PrevIP)
	if err != nil {
		log.Errorf("connect container %s to network %s failed: %s",
			containerId, network, err)
	}
}

func (ch *CalicoHandler) HandleEndpointExist(es ErrorSpec, errMsg string) {
	if EndpointExistError(errMsg) == false {
		return
	}

	netInfoURI := "/networks/" + es.Network
	resp, err := ch.DockerClient.Get("http://unix" + netInfoURI)
	if err != nil {
		log.Errorf("get network %s info failed: %s", es.Network, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	networkInfo := DockerNetwork{}
	json.Unmarshal(body, &networkInfo)
	log.Infof("finish parse network info %s", networkInfo)

	var endpointId string
	networkId := networkInfo.Id
	for _, net := range networkInfo.Containers {
		if net.Name == es.ContainerName {
			endpointId = net.EndpointID
		}
	}

	log.Infof("ready delete endpoint %s in network %s", endpointId, networkId)
	if networkId != "" && endpointId != "" {
		endpointKey := fmt.Sprintf("/docker/network/v1.0/endpoint/%s/%s", networkId, endpointId)
		ch.StorageClient.Remove(endpointKey)
		log.Infof("finish delete endpoint key %s in docker", endpointKey)
	}
}

func (ch *CalicoHandler) HandleIPInUse(es ErrorSpec, errMsg string) {
	IP, shouldRelease := IPInUseError(errMsg)
	if IP == "" || shouldRelease == false {
		return
	}

	payload := []byte(`{"Address":"` + IP + `"}`)
	resp, err := ch.CalicoClient.Post("http://unix/IpamDriver.ReleaseAddress", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Errorf("release IP %s for container %s failed: %s", IP, es.ContainerName, err)
	}
	log.Infof("success release IP %s for container %s, states code %v", IP, es.ContainerName, resp.StatusCode)
}

func findIP(msg string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(msg)
}

func IPInUseError(msg string) (string, bool) {
	regexPattern := "The address (.*) is already in use"
	regEx := regexp.MustCompile(regexPattern)
	if regEx.FindString(msg) != "" {
		return findIP(msg), true
	} else {
		return "", false
	}
}

func EndpointExistError(msg string) bool {
	regexPattern := "service endpoint with name (.*) already exists"
	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(msg) != ""
}

func NoSandboxPresentError(msg string) bool {
	noSandboxMsg := "no sandbox present"
	return strings.Contains(msg, noSandboxMsg)
}

func NewCalicoHandler(cluster cluster.Cluster, store storage.Store) *CalicoHandler {
	cc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", CalicoSockPath)
			},
		},
	}

	dc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", DockerSockPath)
			},
		},
	}

	ch := &CalicoHandler{CalicoClient: cc, DockerClient: dc, StorageClient: store, ClusterClient: cluster}
	return ch
}
