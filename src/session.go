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

func (s *SocksSession) start(client, server *SocksConnection) {

	s.clientConn = client

	if s.clientConn.state != running {
		debug.output("[socks]client connection is offline ")
		return
	}

	s.serverConn = server

	if s.serverConn.state != running {
		debug.output("[socks]server connection is offline ")
		return
	}
	//TODO

}
