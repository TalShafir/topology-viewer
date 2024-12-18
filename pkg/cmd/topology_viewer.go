package cmd

import (
	"context"
	"errors"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
)

type (
	TopologyViewerOptions struct {
		genericiooptions.IOStreams

		configFlags   *genericclioptions.ConfigFlags
		client        kubernetes.Interface
		label         string
		allNamespaces bool
	}

	Toplogy struct {
		Count     int                 `json:"count,omitempty"`
		Resources corev1.ResourceList `json:"resources,omitempty"`
	}
)

// NewTopologyViewerOptions provides an instance of NewTopologyViewerOptions with default values
func NewTopologyViewerOptions(
	client kubernetes.Interface,
	streams genericiooptions.IOStreams, label string,
	configFlags *genericclioptions.ConfigFlags,
	allNamespaces bool,
) *TopologyViewerOptions {
	return &TopologyViewerOptions{
		client:        client,
		IOStreams:     streams,
		label:         label,
		configFlags:   configFlags,
		allNamespaces: allNamespaces,
	}
}

func newTopology() *Toplogy {
	return &Toplogy{
		Resources: corev1.ResourceList{},
	}
}

func mergeResources(sum corev1.ResourceList, adds ...corev1.ResourceList) {
	for _, add := range adds {
		for k, v := range add {
			src, exists := sum[k]
			if !exists {
				src = resource.MustParse("0")
			}

			src.Add(v)
			sum[k] = src
		}
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

func (t *TopologyViewerOptions) Pods(ctx context.Context, labelSelector string) (map[string]*Toplogy, error) {
	pods, nodeNameToTopology, err := t.loadPodsAndNodesToTopology(ctx, labelSelector)
	if err != nil {
		return nil, err
	}

	topologies := make(map[string]*Toplogy, 0)

	for _, pod := range pods {
		// skip not running pods
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		// skip terminating pods
		if pod.DeletionTimestamp != nil {
			continue
		}

		// skip pods without node name
		if pod.Spec.NodeName == "" {
			continue
		}

		topology := "-"
		if t, exist := nodeNameToTopology[pod.Spec.NodeName]; exist {
			topology = t
		}

		t, exists := topologies[topology]
		if !exists {
			t = newTopology()
			topologies[topology] = t
		}

		t.Count++

		podResources := make([]corev1.ResourceList, 0, len(pod.Spec.Containers))
		for _, c := range pod.Spec.Containers {
			podResources = append(podResources, c.Resources.Requests)
		}

		mergeResources(t.Resources, podResources...)
	}

	return topologies, nil
}

func (t *TopologyViewerOptions) loadPodsAndNodesToTopology(ctx context.Context, labelSelector string) ([]corev1.Pod, map[string]string, error) {
	var (
		pods               []corev1.Pod
		nodeNameToTopology map[string]string
	)

	errC := make(chan error, 2)
	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		opts := metav1.ListOptions{}
		if labelSelector != "" {
			opts.LabelSelector = labelSelector
		}

		podsList, err := t.client.CoreV1().Pods(t.Namespace()).List(ctx, opts)
		if err != nil {
			errC <- err
			return
		}

		pods = podsList.Items
	}()

	go func() {
		defer wg.Done()

		nodesList, err := t.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			errC <- err
			return
		}

		nodeNameToTopology = make(map[string]string, len(nodesList.Items))

		for _, node := range nodesList.Items {
			nodeTopology := "-"
			if t, exist := node.GetLabels()[t.label]; exist {
				nodeTopology = t
			}

			nodeNameToTopology[node.Name] = nodeTopology
		}
	}()

	wg.Wait()
	close(errC)

	errs := make([]error, 0, 2)
	for err := range errC {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return nil, nil, errors.Join(errs...)
	}

	return pods, nodeNameToTopology, nil
}

func (t *TopologyViewerOptions) Namespace() string {
	ns, _, err := t.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		panic(err)
	}

	if t.allNamespaces {
		ns = "" // all namespaces
	}

	return ns
}
