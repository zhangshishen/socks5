package socks5

import (
	"net"
)

type SocksSession struct {
	clientConn *SocksConnection
	serverConn *SocksConnection
	clientAddr *net.TCPAddr
	serverAddr *net.TCPAddr
}

func (s *SocksSession) start() {

	if s.clientConn.state != running {
		debug.output("[socks]client connection is offline ")
		return
	}

	if s.serverConn.state != pending {
		debug.output("[socks]server is already running")
		return
	}

}
