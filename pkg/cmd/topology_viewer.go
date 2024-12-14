package cmd

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
)

type (
	TopologyViewerOptions struct {
		genericiooptions.IOStreams

		client kubernetes.Interface
		label  string
	}

	Toplogy struct {
		Count     int                 `json:"count,omitempty"`
		Resources corev1.ResourceList `json:"resources,omitempty"`
	}
)

// NewTopologyViewerOptions provides an instance of NewTopologyViewerOptions with default values
func NewTopologyViewerOptions(client kubernetes.Interface, streams genericiooptions.IOStreams, label string) *TopologyViewerOptions {
	return &TopologyViewerOptions{
		client:    client,
		IOStreams: streams,
		label:     label,
	}
}

func newTopology() *Toplogy {
	return &Toplogy{
		Resources: corev1.ResourceList{},
	}
}

func mergeResources(sum corev1.ResourceList, add corev1.ResourceList) {
	for k, v := range add {
		src, exists := sum[k]
		if !exists {
			src = resource.MustParse("0")
		}

		src.Add(v)
		sum[k] = src
	}
}

func (t *TopologyViewerOptions) Nodes(ctx context.Context) (map[string]*Toplogy, error) {
	nodesList, err := t.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := nodesList.Items

	topologies := make(map[string]*Toplogy, 0)

	for _, node := range nodes {
		topology, exists := node.GetLabels()[t.label]
		if !exists {
			topology = "-"
		}

		t, exists := topologies[topology]
		if !exists {
			t = newTopology()
			topologies[topology] = t
		}

		t.Count++
		mergeResources(t.Resources, node.Status.Allocatable)
	}

	return topologies, nil
}
