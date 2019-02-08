package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
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

func sendHandshakeRequest(conn *net.TCPConn, req *Request) error {
	buf := make([]byte, 1024)
	buf[0] = 5
	buf[1] = 1
	buf[2] = 0
	n, err := conn.Write(buf[0:3])

	if n != 3 || err != nil {
		debug.output("handshake failed, server can't connect")
		return errors.New("handshake failed")
	}

	n, err = conn.Read(buf)

	if n != 2 || err != nil {
		debug.output("handshake failed, server can't connect")
		return errors.New("handshake failed")
	}

	buf[0] = 5
	buf[1] = req.cmd
	buf[2] = 0
	buf[3] = req.atyp
	copy(buf[4:], req.dst_addr)
	buf[4+len(req.dst_addr)] = req.dst_port[0]
	buf[4+len(req.dst_addr)+1] = req.dst_port[1]

	n, err = conn.Write(buf[0 : 4+len(req.dst_addr)+2])

	if err != nil {
		debug.output("handshake failed, server can't connect")
		return errors.New("handshake failed")
	}

	return nil

}

func recvHandshakeRequest(conn *net.TCPConn) (*Request, error) {
	req := &Request{}

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)

	if n < 2 || err != nil {
		return req, errors.New("handshake failed")
	}

	buf[0] = 5
	buf[1] = 0

	n, err = conn.Write(buf[0:2])

	if n != 2 || err != nil {
		return req, errors.New("handshake failed")
	}

	n, err = conn.Read(buf)

	outBuf := buf[0:n]
	fmt.Printf("receive package %d\n", outBuf)
	req.version = outBuf[0]
	req.cmd = outBuf[1]
	req.atyp = outBuf[3]

	if req.atyp == 1 {
		req.dst_addr = append(req.dst_addr, outBuf[4:8]...)
		copy(req.dst_port[:], outBuf[8:10])
	} else if req.atyp == 3 {
		length := outBuf[4]
		req.dst_addr = append(req.dst_addr, outBuf[5:5+length]...)
		copy(req.dst_port[:], outBuf[5+length:7+length])
		//fmt.Printf("port 1 = %d\n", req.dst_port[0])
		//fmt.Printf("port 2 = %d\n", req.dst_port[1])
	} else if req.atyp == 4 {
		req.dst_addr = append(req.dst_addr, outBuf[4:20]...)
		copy(req.dst_port[:], outBuf[20:22])
	}

	return req, nil
}

func sendHandshakeReply(conn *net.TCPConn, r *Reply) error {
	buf := make([]byte, 1024)

	buf[0] = 5
	buf[1] = r.rep
	buf[2] = 0
	buf[3] = 1
	copy(buf[4:], r.bnd_addr)
	buf[4+len(r.bnd_addr)] = r.bnd_port[0]
	buf[4+len(r.bnd_addr)+1] = r.bnd_port[1]

	_, err := conn.Write(buf[0 : 4+len(r.bnd_addr)+2])

	if err != nil {
		debug.output("send handshake reply failed")
		return errors.New("handshake failed")
	}

	return nil
}

func recvHandshakeReply(conn *net.TCPConn) (*Reply, error) {
	rep := &Reply{}

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)

	buf = buf[0:n]

	if err != nil {
		debug.output("recv handshake reply failed")
		return rep, errors.New("handshake failed")
	}

	rep.version = 5
	rep.rep = buf[1]
	rep.atyp = buf[3]

	if rep.atyp == 1 {
		rep.bnd_addr = append(rep.bnd_addr, buf[4:8]...)
		copy(rep.bnd_port[:], buf[8:10])
	} else if rep.atyp == 3 {
		length := buf[4]
		rep.bnd_addr = append(rep.bnd_addr, buf[5:5+length]...)
		copy(rep.bnd_port[:], buf[5+length:7+length])
	} else if rep.atyp == 4 {
		rep.bnd_addr = append(rep.bnd_addr, buf[4:20]...)
		copy(rep.bnd_port[:], buf[20:22])
	}

	return rep, nil
}

func getAddr(r *Request) string {

	if r.atyp == 3 {
		addr := string(r.dst_addr[:])

		b := make([]byte, 4)
		b[0] = 0
		b[1] = 0
		b[2] = r.dst_port[0]
		b[3] = r.dst_port[1]
		Port := int(binary.BigEndian.Uint32(b))

		p := strconv.Itoa(Port)

		return addr + ":" + p

	} else {

		b := make([]byte, 4)
		b[0] = 0
		b[1] = 0
		b[2] = r.dst_port[0]
		b[3] = r.dst_port[1]
		addr := net.TCPAddr{IP: []byte{r.dst_addr[0], r.dst_addr[1], r.dst_addr[2], r.dst_addr[3]}, Port: int(binary.BigEndian.Uint32(b))}
		return addr.String()
	}
}
