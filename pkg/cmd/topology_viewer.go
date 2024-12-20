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
	"k8s.io/klog/v2"
)

type (
	TopologyViewerOptions struct {
		genericiooptions.IOStreams

		configFlags   *genericclioptions.ConfigFlags
		client        kubernetes.Interface
		topologyKey   string
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
	streams genericiooptions.IOStreams,
	topologyKey string,
	configFlags *genericclioptions.ConfigFlags,
	allNamespaces bool,
) *TopologyViewerOptions {
	return &TopologyViewerOptions{
		client:        client,
		IOStreams:     streams,
		topologyKey:   topologyKey,
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
	klog.V(2).InfoS("loading nodes")

	nodesList, err := t.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := nodesList.Items

	klog.V(2).InfoS("loaded nodes", "count", len(nodes))

	topologies := make(map[string]*Toplogy, 0)

	for _, node := range nodes {
		klog.V(4).InfoS("handling node", "node", node.Name)

		topology, exists := node.GetLabels()[t.topologyKey]
		if !exists {
			topology = "-"
		}

		klog.V(4).InfoS("found node topology", "node", node.Name, "topologyName", topology)

		t, exists := topologies[topology]
		if !exists {
			klog.V(4).InfoS("topology doesn't exist creating it", "node", node.Name, "topologyName", topology)

			t = newTopology()
			topologies[topology] = t
		}

		klog.V(4).InfoS("adding node to topology", "node", node.Name, "nodeAllocatable", node.Status.Allocatable, "topologyName", topology, "topology", t)

		t.Count++
		mergeResources(t.Resources, node.Status.Allocatable)

		klog.V(4).InfoS("added node to topology", "node", node.Name, "topologyName", topology, "topology", t)
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
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}

		topology := "-"
		if t, exist := nodeNameToTopology[nodeName]; exist {
			topology = t
		}

		klog.V(4).InfoS("found pod topology", "pod", klog.KObj(&pod), "node", nodeName, "topologyName", topology)

		t, exists := topologies[topology]
		if !exists {
			klog.V(4).InfoS("topology doesn't exist creating it", "pod", klog.KObj(&pod), "node", nodeName, "topologyName", topology)

			t = newTopology()
			topologies[topology] = t
		}

		t.Count++

		podResources := make([]corev1.ResourceList, 0, len(pod.Spec.Containers))
		for _, c := range pod.Spec.Containers {
			podResources = append(podResources, c.Resources.Requests)
		}

		klog.V(4).InfoS("adding node to topology", "node", nodeName, "topologyName", topology, "podResources", podResources, "topology", t)

		mergeResources(t.Resources, podResources...)

		klog.V(4).InfoS("added node to topology", "node", nodeName, "topologyName", topology, "topology", t)
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

	klog.V(2).InfoS("loading pods and nodes")

	go func() {
		defer wg.Done()

		opts := metav1.ListOptions{}
		if labelSelector != "" {
			opts.LabelSelector = labelSelector
		}

		ns := t.Namespace()

		klog.V(2).InfoS("loading pods", "labelSelector", labelSelector, "namespace", ns)

		podsList, err := t.client.CoreV1().Pods(ns).List(ctx, opts)
		if err != nil {
			errC <- err
			return
		}

		pods = podsList.Items

		klog.V(2).InfoS("loaded pods", "labelSelector", labelSelector, "namespace", ns, "count", len(pods))
	}()

	go func() {
		defer wg.Done()

		klog.V(2).InfoS("loading nodes")

		nodesList, err := t.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			errC <- err
			return
		}

		klog.V(2).InfoS("loaded nodes", "count", len(nodesList.Items))

		nodeNameToTopology = make(map[string]string, len(nodesList.Items))

		klog.V(2).InfoS("building node to topology mapping")
		for _, node := range nodesList.Items {
			nodeTopology := "-"
			if t, exist := node.GetLabels()[t.topologyKey]; exist {
				nodeTopology = t
			}

			klog.V(4).InfoS("add node to topology", "node", node.Name, "topology", nodeTopology)

			nodeNameToTopology[node.Name] = nodeTopology
		}

		klog.V(2).InfoS("built node to topology mapping")
		klog.V(4).InfoS("built node to topology mapping", "mapping", nodeNameToTopology)
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
