package multicast

import (
	"context"
	stdlibLog "log"
	"net"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/ravenix/peerd/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type dnsExplorer struct {
	serverConfig *mdns.Config
	clientConfig *mdns.QueryParam
}

type dnsExplorerConfig struct {
	Interface string `yaml:"interface"`
	Hostname  string `yaml:"hostname"`
	Domain    string `yaml:"domain"`
	Service   string `yaml:"service"`
	Port      uint16 `yaml:"port"`
}

func dnsExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config dnsExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newDnsExplorer(&config)
}

func newDnsExplorer(config *dnsExplorerConfig) (*dnsExplorer, error) {
	iface, err := net.InterfaceByName(config.Interface)
	if err != nil {
		return nil, err
	}

	service, err := mdns.NewMDNSService(config.Hostname, config.Service, config.Domain+".", config.Hostname+".", int(config.Port), nil, []string{})
	if err != nil {
		return nil, err
	}

	return &dnsExplorer{
		serverConfig: &mdns.Config{
			Iface: iface,
			Zone:  service,
		},
		clientConfig: &mdns.QueryParam{
			Service:             config.Service,
			Domain:              config.Domain,
			Timeout:             time.Second * 5,
			WantUnicastResponse: false,
			DisableIPv4:         false,
			DisableIPv6:         false,
		},
	}, nil
}

func (e *dnsExplorer) Run(ctx context.Context) error {
	config := *e.serverConfig
	config.Logger = stdlibLog.New(log.StandardLogger().Writer(), "", 0)

	server, err := mdns.NewServer(&config)
	defer server.Shutdown()

	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (e *dnsExplorer) Explore(ctx context.Context, dh explorer.DiscoveryHandler) error {
	entriesCh := make(chan *mdns.ServiceEntry, 256)
	defer close(entriesCh)

	config := *e.clientConfig
	config.Entries = entriesCh
	config.Logger = stdlibLog.New(log.StandardLogger().WriterLevel(log.DebugLevel), "", 0)

	if err := mdns.QueryContext(ctx, &config); err != nil {
		return err
	}

	go func() {
		for entry := range entriesCh {
			dh.Discovered(&explorer.Discovery{
				IPv4Addr: entry.AddrV4,
				IPv6Addr: entry.AddrV6,
				Port:     uint16(entry.Port),
			})
		}
	}()

	<-ctx.Done()
	return nil
}
