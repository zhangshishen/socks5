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

	encrypt EncryptFunc
	decrypt DecryptFunc
}

func defaultEncryptDecript(b []byte) []byte {
	return b
}
func (s *SocksSession) init(id string) {
	s.encrypt = defaultEncryptDecript
	s.decrypt = defaultEncryptDecript
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
	go func() {
		var in []byte
		for {
			select {
			case in = <-client.inQueue:
			case <-client.ctx:
				//client been closed, so close server
				server.close()
				return
			}

			in = s.encrypt(in)

			select {
			case server.outQueue <- in:
			case <-server.ctx:
				return
			}
		}

	}()

	//read from server and send to client
	go func() {
		var in []byte
		for {
			select {
			case in = <-server.inQueue:
			case <-server.ctx:
				//server been closed, so close client
				client.close()
				return
			}

			in = s.decrypt(in)

			select {
			case client.outQueue <- in:
			case <-client.ctx:
				return
			}
		}
	}()
}
