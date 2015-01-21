package main

import (
	"fmt"
	"net"
	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/protobuf"
)

type ScalableLeaderSession struct {
	*Context
	*Broadcaster

	Nonce Nonce

	s_i abstract.Secret
	C_i abstract.Point
	C_i_p []byte

	V_C_p [][]byte
}

func NewScalableLeaderSession(context *Context, nonce Nonce,
		replyConn net.Conn, done chan<- Nonce) chan <-net.Conn {

	broadcaster := &Broadcaster{
		context,
		make([]net.Conn, context.N),
	}

	scalable := &ScalableLeaderSession{
		context,
		broadcaster,
		nonce,
		context.Suite.Secret(),
		context.Suite.Point(),
		nil,
		nil,
	}

	incoming := make(chan net.Conn)
	go scalable.Start(incoming, replyConn, done)

	return incoming
}

func (s *ScalableLeaderSession) Start(connChan <-chan net.Conn,
		replyConn net.Conn, close chan<- Nonce) {

	// Leader connects to everybody else.
	for i := 0; i < s.N; i++ {
		if s.IsMine(i) {
			continue
		}
		conn, err := net.DialTimeout("tcp", s.Peers[i].Addr, timeout)
		if err != nil {
			format := "Unable to connect to server at %s"
			panic(fmt.Sprintf(format, s.Peers[i].Addr))
		}
		buf := protobuf.Encode(&NonceMessage{ s.Nonce })
		if _, err := WritePrefix(conn, buf); err != nil {
			panic("announcement: " + err.Error())
		}
		s.Conns[i] = conn
	}
	fmt.Println("Started Leader " + s.Nonce.String())
	s.RunLottery()
}

func (s *ScalableLeaderSession) GenerateInitialShares() {
	s.s_i.Pick(s.Random)
	s.C_i.Mul(nil, s.s_i)

	h := s.Suite.Hash()
	h.Write(s.C_i.Encode())
	s.C_i_p = h.Sum(nil)
}

func (s *ScalableLeaderSession) RunLottery() {
	s.GenerateInitialShares()

}