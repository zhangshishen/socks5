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
	wc sync.WaitGroup

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
	s.tcpConn.Close()
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

	s.state = running
	return nil
}

func (s *SocksConnection) run(inQueue, outQueue chan []byte) {

	go func(inQueue chan []byte, s *SocksConnection) {
		defer s.wc.Done()
		for {

			n, err := s.tcpConn.Read(s.inBuffer)

			if err != nil {
				close(inQueue)
				return
			}

			s.downstream += uint64(n)
			if n != 0 {
				tmp := make([]byte, n)
				copy(tmp, s.inBuffer)
				inQueue <- tmp
			}

		}

	}(inQueue, s)

	//write routine
	go func(outQueue chan []byte, s *SocksConnection) {
		isClosed := false
		defer s.wc.Done()
		for out := range outQueue {

			if isClosed {
				continue
			}
			n, err := s.tcpConn.Write(out)

			s.upstream += uint64(n)

			if err != nil {
				s.tcpConn.Close()
				isClosed = true
			}
			if n != len(out) {
				debug.output("fatal error, send size less than buffer size")
			}
		}
		s.tcpConn.Close()

	}(outQueue, s)

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
