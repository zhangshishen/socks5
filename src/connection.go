package socks5

import (
	"net"
	"sync"
)

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

const (
	BUFFERSIZE  = 4086
	CHANNELSIZE = 256
)

type SocksConnection struct {
	mu sync.Mutex

	tcpConn *net.TCPConn

	typ       connType
	state     connState
	inQueue   chan []byte
	outQueue  chan []byte
	inBuffer  []byte
	outBuffer []byte
	ctx       chan int
	dst       string //dest address
}

func (s *SocksConnection) init() {

	s.inQueue = make(chan []byte, CHANNELSIZE)
	s.outQueue = make(chan []byte, CHANNELSIZE)
	s.ctx = make(chan int)
	s.inBuffer = make([]byte, BUFFERSIZE)
	s.outBuffer = make([]byte, BUFFERSIZE)

	s.state = pending

}

func (s *SocksConnection) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == dead {
		return
	}

	close(s.ctx)
	close(s.inQueue)
	close(s.outQueue)

	s.tcpConn.Close()
	s.state = dead
}

func (s *SocksConnection) connect(addr string, l *net.TCPListener) error {

	if s.state == pending {
		if l != nil {
			c, err := l.AcceptTCP() //connection from listen
			s.tcpConn = c
			if err != nil {
				debug.output("connect to addr failed")
				return "", err
			}
		} else {
			c, err := net.Dial("tcp", addr)
			s.tcpConn = c.(*net.TCPConn)
			if err != nil {
				debug.output("connect to addr failed")
				return "", err
			}
		}

		s.state = running

	} else {
		debug.output("fatal error , tcpConn is not pending\n")
	}

	return "", nil
}

func (s *SocksConnection) run() {

	go func(c *net.TCPConn, s *SocksConnection) {

		for {
			n, err := c.Read(s.inBuffer)
			if err != nil {
				s.close()
				return
			}
			tmp := make([]byte, n)
			copy(tmp, s.inBuffer)

			select {
			case s.inQueue <- tmp:
			case <-s.ctx:
				return
			}

		}
	}(s.tcpConn, s)
	//write routine
	go func(c *net.TCPConn, s *SocksConnection) {

		for {
			tmp := make([]byte, 0)

			select {
			case tmp = <-s.outQueue:
			case <-s.ctx:
				return
			}

			n, err := c.Write(tmp)

			if err != nil {
				s.close()
				return
			}

			if n != len(tmp) {
				debug.output("fatal error, send size less than buffer size")
			}
		}
	}(s.tcpConn, s)
}
