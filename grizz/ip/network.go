package ip

import (
	"net"

	"github.com/panjf2000/gnet/v2"
)

// NetworkServer is a plain TCP server
// it receives a line-delimited IP address and returns whether it is in the trie
type NetworkServer struct {
	gnet.BuiltinEventEngine

	l    *Looker
	addr string
}

// NewNetworkServer creates a new IPLookerNetworkServer
// given an IPLooker and an address to listen on
// the address format must be one of the following:
// - tcp://listen:port
// - udp://listen:port
// - unix://path
func NewNetworkServer(l *Looker, addr string) *NetworkServer {
	return &NetworkServer{l: l, addr: addr}
}

// OnTraffic is called when a new connection is established
func (ipsrv *NetworkServer) OnTraffic(c gnet.Conn) gnet.Action {
	// Peek up to 46 bytes (max IPv6 length + newline)
	// since the data is always going to be less than 46 bytes
	// we always expect io.ErrShortBuffer to be returned
	buf, _ := c.Next(-1)

	ip := net.ParseIP(string(buf[:len(buf)-1])) // remove newline
	if ip == nil {
		c.Write([]byte("error: invalid IP format\n"))
		return gnet.None
	}

	if ipsrv.l.Contains(ip) {
		c.Write([]byte("1\n"))
	} else {
		c.Write([]byte("0\n"))
	}
	return gnet.None
}

// ListenAndServe starts the server
func (ipsrv *NetworkServer) ListenAndServe() error {
	return gnet.Run(ipsrv, ipsrv.addr, gnet.WithMulticore(true))
}
