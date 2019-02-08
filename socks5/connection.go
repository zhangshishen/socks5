package socks5

import (
	"net"
	"sync"
	"time"
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
	BUFFERSIZE  = 4096
	CHANNELSIZE = 256
)

type SocksConnection struct {
	mu sync.Mutex

	id      string
	tcpConn *net.TCPConn

	typ       connType
	state     connState
	inQueue   chan []byte
	outQueue  chan []byte
	inBuffer  []byte
	outBuffer []byte
	ctx       chan int

	stime      time.Time
	upstream   uint64
	downstream uint64
}

func (s *SocksConnection) init(id string) {

	s.inQueue = make(chan []byte, CHANNELSIZE)
	s.outQueue = make(chan []byte, CHANNELSIZE)
	s.ctx = make(chan int)
	s.inBuffer = make([]byte, BUFFERSIZE)
	s.outBuffer = make([]byte, BUFFERSIZE)

	s.state = pending
	s.stime = time.Now()
	s.id = id

	s.upstream = 0
	s.downstream = 0
}

func (s *SocksConnection) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == dead {
		return
	}

	close(s.ctx)

	s.state = dead

}

func (s *SocksConnection) isClose() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == dead
}

func (s *SocksConnection) connect(addr string) error {

	if s.state != pending {
		debug.output("fatal error , tcpConn is not pending\n")
	}
	debug.out("[socks] start connecting %s \n", addr)
	c, err := net.Dial("tcp", addr)

	if err != nil {
		debug.output("connect to addr failed")
		return err
	}
	debug.out("[socks] connect %s success\n", addr)
	s.tcpConn = c.(*net.TCPConn)

	s.state = running
	return nil
}

func (s *SocksConnection) listen(l *net.TCPListener) error {

	if s.state != pending {
		debug.output("fatal error , tcpConn is not pending\n")
	}
	c, err := l.AcceptTCP()
	s.tcpConn = c
	if err != nil {
		debug.output("connect to addr failed")
		return err
	}
	debug.out("[socks] listen success\n")
	s.state = running
	return nil
}

func (s *SocksConnection) run() {

	go func(c *net.TCPConn, s *SocksConnection) {

		for {
			n, err := c.Read(s.inBuffer)

			debug.out("[socks] %s read %d byte \n", s.id, n)

			s.downstream += uint64(n)

			if err != nil {
				debug.out("[socks] connection closed \n")
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
			s.upstream += uint64(n)
			debug.out("[socks] %s write %d byte \n", s.id, n)
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

func (s *SocksConnection) getSrcAddr() string {
	if s.isClose() {
		return "Closing"
	}
	return ""
}

func (s *SocksConnection) getDstAddr() string {
	if s.isClose() {
		return "Closing"
	}
	return ""
}

func (s *SocksConnection) getRunningTime() string {
	return time.Now().Sub(s.stime).String()
}

func (s *SocksConnection) getUpStream() uint64 {
	return s.upstream
}
func (s *SocksConnection) getDownStream() uint64 {
	return s.downstream
}
