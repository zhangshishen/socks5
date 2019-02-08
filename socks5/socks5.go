package socks5

import (
	"encoding/binary"
	"net"
)

const RELAYSERVER = "192.168.96.87:1080"

type EncryptFunc func(buf []byte) []byte
type DecryptFunc func(buf []byte) []byte

type SocksProxy struct {
	conn *net.TCPConn
}

func (s *SocksProxy) runRelay(host string) {
	addr, err := net.ResolveTCPAddr("tcp", host)
	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()

	if err != nil {
		debug.output("[socks] socks5 listen failed\n")
		return
	}

	for {

		client := &SocksConnection{}
		client.init("client")
		defer client.close()

		err := client.listen(l)
		if err != nil {
			debug.output("[socks] listen failed\n")
			return
		}

		//receive handshake
		req, e := recvHandshakeRequest(client.tcpConn)

		if e != nil {
			debug.output("[socks] handshake error\n")
		}

		server := &SocksConnection{}
		server.init("server")
		defer server.close()

		err = server.connect(RELAYSERVER)
		if err != nil {
			debug.output("[socks] connect to relay server failed\n")
			return
		}

		err = sendHandshakeRequest(server.tcpConn, req)

		if err != nil {
			debug.output("[socks] send to relay server failed\n")
			return
		}

		rep, err := recvHandshakeReply(server.tcpConn)

		if err != nil {
			debug.output("[socks] recv relay reply failed\n")
			return
		}

		err = sendHandshakeReply(client.tcpConn, rep)

		if err != nil {
			debug.output("[socks] recv relay reply failed\n")
			return
		}
		//after handshake, run mainloop(async)
		client.run()
		server.run()

		session := &SocksSession{}
		session.init("")
		go session.start(client, server)

	}
}

func (s *SocksProxy) Run(host string) {
	addr, err := net.ResolveTCPAddr("tcp", host)
	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()

	if err != nil {
		debug.output("[socks] socks5 listen failed\n")
		return
	}

	for {

		client := &SocksConnection{}
		client.init("client")
		defer client.close()

		err := client.listen(l)
		if err != nil {
			debug.output("[socks] listen failed\n")
			return
		}

		debug.out("[socks] listen succeed\n")

		//receive handshake
		req, e := recvHandshakeRequest(client.tcpConn)

		if e != nil {
			debug.output("[socks] handshake error\n")
		}

		debug.out("[socks] handshake succeed\n")

		server := &SocksConnection{}
		server.init("server")
		defer server.close()

		//connect to server,
		debug.out("[socks] resolve address is %s\n", getAddr(req))

		err = server.connect(getAddr(req))

		if err != nil {
			debug.output("[socks] connect to relay server failed\n")
			return
		}

		//UDP method
		if req.cmd == 3 {

		}

		//BIND method
		if req.cmd == 2 {

		}
		rep := &Reply{version: 5, rep: 0, rsv: 0, atyp: req.atyp, bnd_addr: []byte{0, 0, 0, 0}, bnd_port: [2]byte{0, 0}}

		addr := server.tcpConn.LocalAddr()
		a, ok := addr.(*net.TCPAddr)
		if !ok {
			debug.output("[socks] address is not TCP address\n")
		}

		rep.bnd_addr[0] = a.IP[0]
		rep.bnd_addr[1] = a.IP[1]
		rep.bnd_addr[2] = a.IP[2]
		rep.bnd_addr[3] = a.IP[3]

		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs, uint32(a.Port))

		rep.bnd_port[0] = bs[2]
		rep.bnd_port[1] = bs[3]

		debug.out("[socks] start to send reply\n")

		err = sendHandshakeReply(client.tcpConn, rep)

		if err != nil {
			debug.output("[socks] recv relay reply failed\n")
			return
		}
		//after handshake, run mainloop(async)
		debug.out("[socks] start main loop\n")
		client.run()
		server.run()

		session := &SocksSession{}
		session.init("")
		go session.start(client, server)

	}
}
