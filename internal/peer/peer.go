package peer

import (
	"net"
	"time"
)

type Peer struct {
	IPv4Addr net.IP
	IPv6Addr net.IP
	Port     uint16

	FirstSeen time.Time
	LastSeen  time.Time
}
