package socks5

import (
	"net"
)

type VersionRequest struct {
	version byte
	methods []byte
}

type VersionReply struct {
	version byte
	method  byte
}

type Request struct {
	version  byte
	cmd      byte
	rsv      byte
	atyp     byte
	dst_addr []byte
	dst_port [2]byte
}

type Reply struct {
	version  byte
	rep      byte
	rsv      byte
	atyp     byte
	bnd_addr []byte
	bnd_port [2]byte
}

func handshake(conn *net.TCPConn) (*SocksConnection, error) {
	res := new(SocksConnection)

	res.init(conn)
	return res, nil
}
