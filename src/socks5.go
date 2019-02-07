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

		client := &SocksConnection{}
		client.init()
		err := client.connect("", l)
		//client, err := handshake(conn)

		if err != nil {
			debug.output("[socks] listen failed\n")
			return
		}
		client.connect()
		session := new(SocksSession)
		session.start()

		go session.start(client)
	}
}
