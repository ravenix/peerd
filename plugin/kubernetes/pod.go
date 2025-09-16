package kubernetes

import (
	"context"
	"net"

	"github.com/ravenix/peerd/pkg/explorer"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type podExplorer struct {
	*resourceExplorer
	namespace string
	podPort   uint16
}

type podExplorerConfig struct {
	resourceExplorerConfig `yaml:",inline"`
	Namespace              string `yaml:"namespace"`
	PodPort                uint16 `yaml:"pod_port"`
}

func podExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config podExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newpodExplorer(&config)
}

func newpodExplorer(config *podExplorerConfig) (*podExplorer, error) {
	e := &podExplorer{}
	arc, err := newResourceExplorer(&config.resourceExplorerConfig, e.listPods, e.explorePod)

	if err != nil {
		return nil, err
	}

	e.resourceExplorer = arc
	e.podPort = config.PodPort
	return e, nil
}

func (e *podExplorer) listPods(ctx context.Context) ([]any, error) {
	list, err := e.k8sClient.CoreV1().Pods(e.namespace).List(ctx, e.newListOptions())
	if err != nil {
		return nil, err
	}

	var items []any
	for _, item := range list.Items {
		items = append(items, item)
	}

	return items, nil
}

func (e *podExplorer) explorePod(ctx context.Context, resource any) *explorer.Discovery {
	pod, ok := resource.(corev1.Pod)
	if !ok {
		return nil
	}

	if ipAddr := net.ParseIP(pod.Status.PodIP); ipAddr != nil {
		dis := &explorer.Discovery{
			Port: e.podPort,
		}

		if ipAddr.To4() != nil {
			dis.IPv4Addr = ipAddr
		} else {
			dis.IPv6Addr = ipAddr
		}

		return dis
	}

	return nil
}
