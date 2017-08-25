package k8

import (
	"regexp"
	"testing"

	"github.com/wearefair/log-aggregator/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/pkg/api/v1"
)

func TestMatchRegex(t *testing.T) {
	regex, err := regexp.Compile(KubernetesContainerNameRegexp)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		format   string
		expected map[string]string
	}{
		// Kubernetes 1.4 format
		{
			format: "k8s_contract-service.165099f9_contract-service-846522753-4ipvq_default_fb181b9e-5617-11e7-9ea7-06dbc747ad30_87a62e46",
			expected: map[string]string{
				"container_name": "contract-service",
				"pod_name":       "contract-service-846522753-4ipvq",
				"namespace":      "default",
			},
		},

		// Kubernetes 1.6 format
		{
			format: "k8s_public-api_public-api-962190421-dr91z_default_aeac8a6e-5ad9-11e7-8f60-0250a47643e4_0",
			expected: map[string]string{
				"container_name": "public-api",
				"pod_name":       "public-api-962190421-dr91z",
				"namespace":      "default",
			},
		},
	}

	for index, testCase := range testCases {
		result := matchRegex(testCase.format, regex)
		for k, v := range testCase.expected {
			if v != result[k] {
				t.Errorf("Expected %s in test case %d to be %s, but got %s", k, index, v, result[k])
			}
		}
	}
}

func TestTransform(t *testing.T) {
	pods := make(map[string]*v1.Pod)
	track := &mockTracker{
		pods: pods,
	}
	k8 := NewWithTracker(track, Config{})

	// Test record without being able to find pod metadata
	rec := &types.Record{
		Fields: map[string]interface{}{
			"CONTAINER_NAME":    "k8s_containername.containerhash_podname_namespacename_poduuid_abcd1234",
			"CONTAINER_ID_FULL": "mycontainerid",
		},
	}
	transformed, _ := k8.Transform(rec)
	if val := transformed.Fields["docker"].(metadataDocker).ContainerId; val != "mycontainerid" {
		t.Errorf("Expected container id to be %s, but got %s", "mycontainerid", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).NamespaceName; val != "namespacename" {
		t.Errorf("Expected NamespaceName to be %s, but got %s", "namespacename", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodName; val != "podname" {
		t.Errorf("Expected PodName to be %s, but got %s", "podname", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).ContainerName; val != "containername" {
		t.Errorf("Expected ContainerName to be %s, but got %s", "contianername", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodId; val != "" {
		t.Errorf("Expected PodId to be empty, but got %s", val)
	}

	// Test record with pod metadata
	rec = &types.Record{
		Fields: map[string]interface{}{
			"CONTAINER_NAME":    "k8s_containername.containerhash_podname_namespacename_poduuid_abcd1234",
			"CONTAINER_ID_FULL": "mycontainerid",
		},
	}
	pods["namespacename_podname"] = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: k8types.UID("poduid"),
			Labels: map[string]string{
				"label1": "value1",
			},
		},
		Spec: v1.PodSpec{
			NodeName: "myhost",
		},
	}
	transformed, _ = k8.Transform(rec)
	if val := transformed.Fields["docker"].(metadataDocker).ContainerId; val != "mycontainerid" {
		t.Errorf("Expected container id to be %s, but got %s", "mycontainerid", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).NamespaceName; val != "namespacename" {
		t.Errorf("Expected NamespaceName to be %s, but got %s", "namespacename", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodName; val != "podname" {
		t.Errorf("Expected PodName to be %s, but got %s", "podname", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).ContainerName; val != "containername" {
		t.Errorf("Expected ContainerName to be %s, but got %s", "contianername", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodId; val != "poduid" {
		t.Errorf("Expected PodId to be poduid, but got %s", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).Node; val != "myhost" {
		t.Errorf("Expected Node to be myhost, but got %s", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).Labels["label1"]; val != "value1" {
		t.Errorf("Expected Labels.label1 to be value1, but got %s", val)
	}

	// Test 1.6 record with pod metadata
	rec = &types.Record{
		Fields: map[string]interface{}{
			"CONTAINER_NAME":    "k8s_my-nginx_my-nginx-379829228-gb3mv_default_3341b837-5b59-11e7-b5c3-024de65267be_0",
			"CONTAINER_ID_FULL": "mycontainerid",
		},
	}
	pods["default_my-nginx-379829228-gb3mv"] = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: k8types.UID("poduid"),
			Labels: map[string]string{
				"label1": "value1",
			},
		},
		Spec: v1.PodSpec{
			NodeName: "myhost",
		},
	}
	transformed, _ = k8.Transform(rec)
	if val := transformed.Fields["docker"].(metadataDocker).ContainerId; val != "mycontainerid" {
		t.Errorf("Expected container id to be %s, but got %s", "mycontainerid", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).NamespaceName; val != "default" {
		t.Errorf("Expected NamespaceName to be %s, but got %s", "default", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodName; val != "my-nginx-379829228-gb3mv" {
		t.Errorf("Expected PodName to be %s, but got %s", "my-nginx-379829228-gb3mv", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).ContainerName; val != "my-nginx" {
		t.Errorf("Expected ContainerName to be %s, but got %s", "my-nginx", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).PodId; val != "poduid" {
		t.Errorf("Expected PodId to be poduid, but got %s", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).Node; val != "myhost" {
		t.Errorf("Expected Node to be myhost, but got %s", val)
	}
	if val := transformed.Fields["kubernetes"].(metadataKubernetes).Labels["label1"]; val != "value1" {
		t.Errorf("Expected Labels.label1 to be value1, but got %s", val)
	}
}

type mockTracker struct {
	pods map[string]*v1.Pod
}

func (t *mockTracker) Get(namespaceName, podName string) *v1.Pod {
	if pod, ok := t.pods[namespaceName+"_"+podName]; ok {
		return pod
	}
	return nil
}
