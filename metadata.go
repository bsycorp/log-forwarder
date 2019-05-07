package main

import (
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	kSumologicCategoryLabel = "com.sumologic/sourceCategory"
	kSumologicSourceLabel   = "com.sumologic/sourceName"

	kKubernetesPodName                = "io.kubernetes.pod.name"
	kKubernetesPodNamespace           = "io.kubernetes.pod.namespace"
	kKubernetesSourceCategoryOverride = "annotation.io." + kSumologicCategoryLabel
	kKubernetesSourceNameOverride     = "annotation.io." + kSumologicSourceLabel

	kContainerTrustedTimestampName = "com.sumologic/trusted-timestamp"
)

var dockerClient *docker.Client

var defaultMetadataValues MetadataValues

type MetadataValues struct {
	source           string
	category         string
	host             string
	trustedTimestamp bool
}

func SetMetadataDefaults(defaults MetadataValues) {
	defaultMetadataValues = defaults
}

func GetMetadataDefaults() MetadataValues {
	return defaultMetadataValues
}

func GetMetadataForProcess(categoryName string, processName string) (values MetadataValues) {
	return MetadataValues{
		category:         defaultMetadataValues.category + "/" + categoryName + "/" + processName,
		host:             defaultMetadataValues.host,
		source:           processName,
		trustedTimestamp: true,
	}
}

func GetMetadataForContainerID(fullContainerID string) (values MetadataValues) {
	//find container detail by id from docker host
	container, err := getDockerContainerInfo(fullContainerID)
	if err != nil || container == nil {
		log.Print("Error getting container info", err)
		return defaultMetadataValues
	}

	containerName := container.Names[0]
	if strings.HasPrefix(containerName, "/") {
		//trim leading slash, is ugly.
		containerName = containerName[1:]
	}

	metadata := MetadataValues{
		category:         defaultMetadataValues.category + "/docker/" + containerName,
		host:             defaultMetadataValues.host,
		source:           containerName,
		trustedTimestamp: false, //default to being untrusted as label/annotation will flag its trusted
	}

	if strings.HasPrefix(containerName, "k8s_") {
		//default pod owner name to pod name, some pods don't have an 'owner'
		podOwnerName := container.Labels[kKubernetesPodName]

		pod, err := getKubernetesPodInfo(fullContainerID)
		if err != nil || pod == nil {
			//error, continue and just use docker derived values
			log.Println("Error getting pod info", err)
		} else {
			if len(pod.Metadata.OwnerReferences) > 0 {
				podOwnerName = pod.Metadata.OwnerReferences[0].Name
			}
		}

		//is kube so get metadata from kube labels / annotations
		metadata.category = defaultMetadataValues.category + "/kubernetes/" + container.Labels[kKubernetesPodNamespace] + "/" + podOwnerName
		metadata.source = container.Labels[kKubernetesPodNamespace] + "." + container.Labels[kKubernetesPodName]

		if len(container.Labels[kKubernetesSourceCategoryOverride]) > 0 {
			metadata.category = defaultMetadataValues.category + "/kubernetes/" + container.Labels[kKubernetesSourceCategoryOverride]
		}
		if len(container.Labels[kKubernetesSourceNameOverride]) > 0 {
			metadata.source = container.Labels[kKubernetesPodNamespace] + "." + container.Labels[kKubernetesSourceNameOverride]
		}
	} else {
		//is docker so get from docker labels assuming no kube, override if found via labels
		if len(container.Labels[kSumologicCategoryLabel]) > 0 {
			metadata.category = defaultMetadataValues.category + "/docker/" + container.Labels[kSumologicCategoryLabel]
		}
		if len(container.Labels[kSumologicSourceLabel]) > 0 {
			metadata.source = container.Labels[kSumologicSourceLabel]
		}
		if MapKeysContains(container.Labels, kContainerTrustedTimestampName) {
			metadata.trustedTimestamp = true
		}
	}

	return metadata
}

func getDockerClient() (*docker.Client, error) {
	if dockerClient == nil {
		var err error
		endpoint := "unix:///var/run/docker.sock"
		dockerClient, err = docker.NewClient(endpoint)
		if err != nil {
			return nil, err
		}
	}
	return dockerClient, nil
}

func getDockerContainerInfo(fullContainerID string) (*docker.APIContainers, error) {
	client, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	//find container detail by id from docker host
	containers, err := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"id": {fullContainerID}}})
	//could be error, or container might have been killed by the time we check
	if err != nil || len(containers) == 0 || len(containers[0].Names) == 0 {
		return nil, err
	}

	return &containers[0], nil
}

//lookup pod list from local kubelet, find more info about pod including its real owner
func getKubernetesPodInfo(fullContainerID string) (*Pod, error) {
	podListResponse, err := http.Get("http://127.0.0.1:10255/pods")
	if err != nil {
		return nil, err
	}

	podListBody, err := ioutil.ReadAll(podListResponse.Body)
	_ = podListResponse.Body.Close()
	if err != nil {
		return nil, err
	}

	var result Pod
	var podListObj PodList
	json.Unmarshal(podListBody, &podListObj)

OuterLoop:
	for _, pod := range podListObj.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if strings.Contains(container.ContainerID, fullContainerID) && len(pod.Metadata.OwnerReferences) > 0 {
				//found matching container, set podOwner to be referenced owner (daemonset, deployment etc name)
				result = pod
				break OuterLoop
			}
		}
	}

	return &result, nil
}

func GetHostname(hostnameFromEnv string) string {
	if hostnameFromEnv != "" {
		return hostnameFromEnv
	}

	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	//try AWS metadata first
	hostnameResponse, err := client.Get("http://169.254.169.254/latest/meta-data/hostname")
	if err != nil || hostnameResponse.StatusCode != 200 {
		log.Println("Error getting AWS hostname:", err)
	} else {
		defer hostnameResponse.Body.Close()
		hostnameBody, _ := ioutil.ReadAll(hostnameResponse.Body)
		return string(hostnameBody)
	}

	//try GCP metadata
	gcpRequest, _ := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/hostname", nil)
	gcpRequest.Header.Add("Metadata-Flavor", "Google")
	hostnameResponse, err = client.Do(gcpRequest)
	if err != nil || hostnameResponse.StatusCode != 200 {
		log.Println("Error getting GCP hostname:", err)
	} else {
		defer hostnameResponse.Body.Close()
		hostnameBody, _ := ioutil.ReadAll(hostnameResponse.Body)
		return string(hostnameBody)
	}

	//then fallback to /etc/hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Error getting hostname:", err)
		return "unknown"
	}

	return hostname
}
