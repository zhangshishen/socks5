package socks5

import (
	"net"
)

type SocksSession struct {
	id string

	clientConn *SocksConnection
	serverConn *SocksConnection

	clientAddr *net.TCPAddr
	serverAddr *net.TCPAddr

	inQueue  chan []byte
	outQueue chan []byte
	encrypt  EncryptFunc
	decrypt  DecryptFunc
}

func defaultEncryptDecript(b []byte) []byte {
	return b
}
func (s *SocksSession) init(id string) {
	s.encrypt = defaultEncryptDecript
	s.decrypt = defaultEncryptDecript

	s.inQueue = make(chan []byte, 64)
	s.outQueue = make(chan []byte, 64)
}

func (s *SocksSession) setEncrypt(e EncryptFunc) {
	s.encrypt = e

}

func (s *SocksSession) setDecrypt(d DecryptFunc) {
	s.decrypt = d

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
	//read from client and send to server
	client.run(s.inQueue, s.outQueue)
	server.run(s.outQueue, s.inQueue)
}
