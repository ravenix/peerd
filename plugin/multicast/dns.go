package multicast

import (
	"context"
	"fmt"
	stdlibLog "log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
	"github.com/ravenix/peerd/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type dnsExplorer struct {
	instanceId    string
	serviceDomain string
	serverConfig  *mdns.Config
	clientConfig  *mdns.QueryParam
}

type dnsExplorerConfig struct {
	InstanceId string   `yaml:"instance_id"`
	Interface  string   `yaml:"interface"`
	IPs        []net.IP `yaml:"ips"`
	IPFilter   []string `yaml:"allowed_ips"`
	Hostname   string   `yaml:"hostname"`
	Domain     string   `yaml:"domain"`
	Service    string   `yaml:"service"`
	Port       uint16   `yaml:"port"`
}

func dnsExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config dnsExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newDnsExplorer(&config)
}

func newDnsExplorer(config *dnsExplorerConfig) (*dnsExplorer, error) {
	var ipFilters []net.IPNet
	for _, ipFilterStr := range config.IPFilter {
		_, ipFilter, err := net.ParseCIDR(ipFilterStr)

		if err != nil {
			return nil, err
		}

		ipFilters = append(ipFilters, *ipFilter)
	}

	iface, err := net.InterfaceByName(config.Interface)
	if err != nil {
		return nil, err
	}

	ifaceAddrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var ifaceIPs []net.IP
	if len(config.IPs) > 0 {
		ifaceIPs = config.IPs
	} else {
		for _, addr := range ifaceAddrs {
			switch v := addr.(type) {
			case *net.IPAddr:
				ifaceIPs = append(ifaceIPs, v.IP)
			case *net.IPNet:
				ifaceIPs = append(ifaceIPs, v.IP)
			}
		}
	}

	var serviceIPs []net.IP
	if len(config.IPFilter) > 0 {
		for _, ip := range ifaceIPs {
			log.Debugf("ip address %s", ip)
			for _, ipFilter := range ipFilters {
				if ipFilter.Contains(ip) {
					log.Debugf("suitable ip address %s", ip)
					serviceIPs = append(serviceIPs, ip)
				}
			}
		}
	} else {
		serviceIPs = ifaceIPs
	}

	if len(serviceIPs) == 0 {
		return nil, fmt.Errorf("no suitable IP addresses for interface %s", config.Interface)
	}

	var instanceId string

	if config.InstanceId != "" {
		instanceId = config.InstanceId
	} else {
		instanceUUID, err := uuid.NewRandom()

		if err != nil {
			return nil, err
		}

		instanceId = instanceUUID.String()
	}

	service, err := mdns.NewMDNSService(
		instanceId,
		config.Service,
		config.Domain+".",
		config.Hostname+".",
		int(config.Port),
		ifaceIPs,
		[]string{},
	)
	if err != nil {
		return nil, err
	}

	return &dnsExplorer{
		instanceId:    instanceId,
		serviceDomain: config.Service + "." + config.Domain + ".",
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
			if entry.Name == e.instanceId+"."+e.serviceDomain {
				log.Debugf("skipping entry %s as it's ourself", entry.Name)
				continue
			}

			if !strings.HasSuffix(entry.Name, "."+e.serviceDomain) {
				log.Debugf("skipping alien service %s", entry.Name)
				continue
			}

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
