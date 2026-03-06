package keepalived

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/ravenix/peerd/pkg/explorer"
	"gopkg.in/yaml.v3"
)

const (
	keepalivedServiceName      = "org.keepalived.Vrrp1"
	keepalivedInstanceIfcName  = "org.keepalived.Vrrp1.Instance"
	keepalivedInstanceProperty = "State"
	keepalivedMasterState      = 2
)

type instanceExplorer struct {
	instanceObjectPath string
	peerIPv4           net.IP
	peerIPv6           net.IP
	port               uint16
	readState          func(context.Context, string) (int64, error)
}

type instanceExplorerConfig struct {
	Interface       string `yaml:"interface"`
	VirtualRouterID uint16 `yaml:"virtual_router_id"`
	PeerIPv4        string `yaml:"peer_ipv4"`
	PeerIPv6        string `yaml:"peer_ipv6"`
	Port            uint16 `yaml:"port"`
}

func instanceExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config instanceExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newInstanceExplorer(&config)
}

func newInstanceExplorer(config *instanceExplorerConfig) (*instanceExplorer, error) {
	if config.Interface == "" {
		return nil, fmt.Errorf("interface must be set")
	}

	if config.VirtualRouterID == 0 {
		return nil, fmt.Errorf("virtual_router_id must be set")
	}

	peerIPv4, err := parseIPv4(config.PeerIPv4)
	if err != nil {
		return nil, fmt.Errorf("invalid peer_ipv4: %w", err)
	}

	peerIPv6, err := parseIPv6(config.PeerIPv6)
	if err != nil {
		return nil, fmt.Errorf("invalid peer_ipv6: %w", err)
	}

	return &instanceExplorer{
		instanceObjectPath: instanceObjectPath(config.Interface, config.VirtualRouterID),
		peerIPv4:           peerIPv4,
		peerIPv6:           peerIPv6,
		port:               config.Port,
		readState:          readStateFromDBus,
	}, nil
}

func (e *instanceExplorer) Run(ctx context.Context) error {
	return nil
}

func (e *instanceExplorer) Cadence() explorer.Cadence {
	return explorer.Cadence{
		ExploreInterval: 100 * time.Millisecond,
		ExploreTimeout:  80 * time.Millisecond,
		PeerTTL:         300 * time.Millisecond,
	}
}

func (e *instanceExplorer) Explore(ctx context.Context, dh explorer.DiscoveryHandler) error {
	state, err := e.readState(ctx, e.instanceObjectPath)
	if err != nil {
		return err
	}

	if state != keepalivedMasterState {
		return nil
	}

	dh.Discovered(&explorer.Discovery{
		IPv4Addr: copyIP(e.peerIPv4),
		IPv6Addr: copyIP(e.peerIPv6),
		Port:     e.port,
	})

	return nil
}

func instanceObjectPath(iface string, virtualRouterID uint16) string {
	return fmt.Sprintf("/org/keepalived/Vrrp1/Instance/%s/%d/IPv6", iface, virtualRouterID)
}

func readStateFromDBus(ctx context.Context, objectPath string) (int64, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	obj := conn.Object(keepalivedServiceName, dbus.ObjectPath(objectPath))

	call := obj.CallWithContext(
		ctx,
		"org.freedesktop.DBus.Properties.Get",
		0,
		keepalivedInstanceIfcName,
		keepalivedInstanceProperty,
	)
	if call.Err != nil {
		return 0, call.Err
	}

	var value dbus.Variant
	if err := call.Store(&value); err != nil {
		return 0, err
	}

	return stateAsInt64(value.Value())
}

func stateAsInt64(value any) (int64, error) {
	if state, ok, err := numericAsInt64(value); ok || err != nil {
		return state, err
	}

	switch v := value.(type) {
	case dbus.Variant:
		return stateAsInt64(v.Value())
	case []interface{}:
		return stateAsInt64FromSlice(v)
	case []dbus.Variant:
		values := make([]interface{}, 0, len(v))
		for _, item := range v {
			values = append(values, item)
		}
		return stateAsInt64FromSlice(values)
	case string:
		return stateAsInt64FromString(v)
	default:
		return 0, fmt.Errorf("unexpected state type %T", value)
	}
}

func numericAsInt64(value any) (int64, bool, error) {
	switch v := value.(type) {
	case int:
		return int64(v), true, nil
	case int8:
		return int64(v), true, nil
	case int16:
		return int64(v), true, nil
	case int32:
		return int64(v), true, nil
	case int64:
		return v, true, nil
	case uint8:
		return int64(v), true, nil
	case uint16:
		return int64(v), true, nil
	case uint32:
		return int64(v), true, nil
	case uint64:
		if v > uint64(math.MaxInt64) {
			return 0, true, fmt.Errorf("state value overflows int64")
		}
		return int64(v), true, nil
	default:
		return 0, false, nil
	}
}

func stateAsInt64FromSlice(values []interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("unexpected state type []interface {}")
	}

	var firstErr error
	for _, value := range values {
		state, err := stateAsInt64(value)
		if err == nil {
			return state, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return 0, firstErr
	}

	return 0, fmt.Errorf("unexpected state type []interface {}")
}

func stateAsInt64FromString(value string) (int64, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "INIT":
		return 0, nil
	case "BACKUP":
		return 1, nil
	case "MASTER":
		return keepalivedMasterState, nil
	case "FAULT", "STOP", "DELETED":
		return 3, nil
	default:
		return 0, fmt.Errorf("unexpected state string %q", value)
	}
}

func parseIPv4(value string) (net.IP, error) {
	if value == "" {
		return nil, nil
	}

	ipAddr := net.ParseIP(value)
	if ipAddr == nil || ipAddr.To4() == nil {
		return nil, fmt.Errorf("expected valid IPv4 address")
	}

	return copyIP(ipAddr.To4()), nil
}

func parseIPv6(value string) (net.IP, error) {
	if value == "" {
		return nil, nil
	}

	ipAddr := net.ParseIP(value)
	if ipAddr == nil || ipAddr.To4() != nil {
		return nil, fmt.Errorf("expected valid IPv6 address")
	}

	return copyIP(ipAddr.To16()), nil
}

func copyIP(ipAddr net.IP) net.IP {
	if ipAddr == nil {
		return nil
	}

	tmp := make(net.IP, len(ipAddr))
	copy(tmp, ipAddr)
	return tmp
}
