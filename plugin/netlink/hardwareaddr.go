package netlink

import (
	"context"
	"net"

	"github.com/ravenix/peerd/pkg/explorer"
	"gopkg.in/yaml.v3"
)

type hardwareAddrExplorer struct {
	iface string
}

type hardwareAddrExplorerConfig struct {
	Interface string `yaml:"interface"`
}

func hardwareAddrExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config hardwareAddrExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newHardwareAddrExplorer(&config)
}

func newHardwareAddrExplorer(config *hardwareAddrExplorerConfig) (*hardwareAddrExplorer, error) {
	return &hardwareAddrExplorer{
		iface: config.Interface,
	}, nil
}

func (e *hardwareAddrExplorer) Run(ctx context.Context) error {
	return nil
}

func (e *hardwareAddrExplorer) Explore(ctx context.Context, dh explorer.DiscoveryHandler) error {
	iface, err := net.InterfaceByName(e.iface)
	if err != nil {
		return err
	}

	hardwareAddr := iface.HardwareAddr
	ipv4Addr := make(net.IP, 4)
	ipv6Addr := make(net.IP, 16)

	copy(ipv4Addr[:], hardwareAddr[len(hardwareAddr)-len(ipv4Addr):])
	copy(ipv6Addr[len(ipv6Addr)-len(hardwareAddr):], hardwareAddr[:])

	dh.Discovered(&explorer.Discovery{
		IPv4Addr: ipv4Addr,
		IPv6Addr: ipv6Addr,
	})

	return nil
}
