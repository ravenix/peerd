package kubernetes

import (
	"context"
	"net"

	"github.com/ravenix/peerd/pkg/explorer"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type nodeExplorer struct {
	*resourceExplorer
	addressType corev1.NodeAddressType
	nodePort    uint16
}

type nodeExplorerConfig struct {
	resourceExplorerConfig `yaml:",inline"`
	AddressType            corev1.NodeAddressType
	NodePort               uint16 `yaml:"node_port"`
}

func nodeExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config nodeExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newNodeExplorer(&config)
}

func newNodeExplorer(config *nodeExplorerConfig) (*nodeExplorer, error) {
	e := &nodeExplorer{}
	arc, err := newResourceExplorer(&config.resourceExplorerConfig, e.listNodes, e.exploreNode)

	if err != nil {
		return nil, err
	}

	e.resourceExplorer = arc

	if config.AddressType != "" {
		e.addressType = config.AddressType
	} else {
		e.addressType = corev1.NodeInternalIP
	}

	e.nodePort = config.NodePort
	return e, nil
}

func (e *nodeExplorer) listNodes(ctx context.Context) ([]any, error) {
	list, err := e.k8sClient.CoreV1().Nodes().List(ctx, e.newListOptions())
	if err != nil {
		return nil, err
	}

	var items []any
	for _, item := range list.Items {
		items = append(items, item)
	}

	return items, nil
}

func (e *nodeExplorer) exploreNode(ctx context.Context, resource any) *explorer.Discovery {
	node, ok := resource.(corev1.Node)
	if !ok {
		return nil
	}

	for _, nodeAddr := range node.Status.Addresses {
		if ipAddr := net.ParseIP(nodeAddr.Address); ipAddr != nil && nodeAddr.Type == e.addressType {
			dis := &explorer.Discovery{
				Port: e.nodePort,
			}

			if ipAddr.To4() != nil {
				dis.IPv4Addr = ipAddr
			} else {
				dis.IPv6Addr = ipAddr
			}

			return dis
		}
	}

	return nil
}
