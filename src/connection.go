package socks5

import "net"

type connType int
type connState int

const (
	tcp = iota
	udp
)

const (
	running = iota
	pending
	initial
	dead
)

type SocksConnection struct {
	conn     net.TCPConn
	typ      connType
	state    connState
	inQueue  chan []byte
	outQueue chan []byte
	inBuffer []byte
	outBuffer []byte
}

func (s *SocksConnection) init() {
	s.inQueue = make(chan []byte,256)
	s.outQueue = make(chan []byte,256)
}


func (s *SocksConnection) connect(addr *net.TCPAddr) error {
	if s.conn != nil {
		debug.output("[socks] already has connection\n")
		return
	}
	s.conn,err := net.DialTCP("tcp",addr)
	if err != nil {
		debug.output("connect to addr failed")
	}

	go func(c net.TCPConn){
		
		for　{
			n,err := c.Read(s.inBuffer)
			if err!= nil {
				c.Close()
			}
			tmp := make([]byte,n)
			append(tmp[0:],s.inBuffer[0:])
		}
	}(s.conn)
}
