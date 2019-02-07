package socks5

import (
	"net"
)

type SocksProxy struct {
	conn *net.TCPConn
}

func (s *SocksProxy) run(host string) {
	addr, err := net.ResolveTCPAddr("tcp", host)
	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()

	if err != nil {
		debug.output("[socks] socks5 listen failed\n")
		return
	}

	for {
		conn, err := l.AcceptTCP()

		if err != nil {
			debug.output("[socks] socks5 accept failed\n")
			return
		}

		session, err := handshake(conn)

		if err != nil {
			debug.output("[socks] socks5 handshake failed\n")
			return
		}

		go session.start()
	}
}
