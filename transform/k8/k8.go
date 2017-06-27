package k8

import (
	"regexp"

	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/types"
)

const (
	CONTAINER_NAME                = "CONTAINER_NAME"
	CONTAINER_ID_FULL             = "CONTAINER_ID_FULL"
	KubernetesContainerNameRegexp = `^k8s_(?P<container_name>[^\._]+)\.?[^_]*_(?P<pod_name>[^_]+)_(?P<namespace>[^_]+)_[^_]+_[a-f0-9]+$`
)

type Config struct {
	KubernetesContainerNameRegexp string
	K8ConfigPath                  string
	NodeName                      string
	MaxPodsCache                  int
}

type Client struct {
	containerNameRegex *regexp.Regexp
	tracker            tracker
}

func New(conf Config) *Client {
	k8, err := newK8(conf.K8ConfigPath)
	if err != nil {
		panic(err)
	}

	tracker := newPodTracker(k8, conf.NodeName, conf.MaxPodsCache)
	tracker.watchForPods()
	return NewWithTracker(tracker, conf)
}

func NewWithTracker(tracker tracker, conf Config) *Client {
	regex := KubernetesContainerNameRegexp
	if conf.KubernetesContainerNameRegexp != "" {
		regex = conf.KubernetesContainerNameRegexp
	}

	compiled, err := regexp.Compile(regex)
	if err != nil {
		panic(errors.Wrapf(err, "Error compiling kubernetes container name regex: %s", regex))
	}

	return &Client{
		containerNameRegex: compiled,
		tracker:            tracker,
	}
}

func (c *Client) Transform(rec *types.Record) (*types.Record, error) {
	containerName, namePresent := rec.Fields[CONTAINER_NAME]
	containerId, idPresent := rec.Fields[CONTAINER_ID_FULL]

	if namePresent && idPresent {

		matchFields := matchRegex(containerName.(string), c.containerNameRegex)
		if matchFields != nil {
			rec.Fields["docker"] = metadataDocker{
				ContainerId: containerId.(string),
			}

			var metadata metadataKubernetes

			if val, ok := matchFields["namespace"]; ok {
				metadata.NamespaceName = val
			}
			if val, ok := matchFields["pod_name"]; ok {
				metadata.PodName = val
			}
			if val, ok := matchFields["container_name"]; ok {
				metadata.ContainerName = val
			}

			pod := c.tracker.Get(metadata.NamespaceName, metadata.PodName)
			if pod != nil {
				metadata.PodId = string(pod.ObjectMeta.UID)
				metadata.Labels = pod.ObjectMeta.Labels
				metadata.Node = pod.Spec.NodeName
			}

			rec.Fields["kubernetes"] = metadata
		}
	}
	return rec, nil
}

func matchRegex(input string, regex *regexp.Regexp) map[string]string {
	match := regex.FindStringSubmatch(input)
	if match == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}
	return result
}

type metadataDocker struct {
	ContainerId string `json:"container_id,omitempty"`
}

type metadataKubernetes struct {
	NamespaceName string            `json:"namespace_name,omitempty"`
	NamespaceId   string            `json:"namespace_id,omitempty"`
	PodName       string            `json:"pod_name,omitempty"`
	PodId         string            `json:"pod_id,omitempty"`
	ContainerName string            `json:"container_name,omitempty"`
	Node          string            `json:"node,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}
