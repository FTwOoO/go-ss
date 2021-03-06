package connection

import (
	"fmt"
	"github.com/FTwOoO/go-ss/dialer"
	"github.com/FTwOoO/go-ss/socks"
	"log"
	"net"
	"time"
)

type ShadowsocksRawConnParams struct {
	Target   socks.Addr
	IsServer bool
}

var _ dialer.ForwardConnection = &ShadowsocksRawConn{}

type ShadowsocksRawConn struct {
	Conn                net.Conn
	params              ShadowsocksRawConnParams
	isServerTargetRead  bool
	isClientTargetWrite bool
	forwardReady        chan socks.Addr
}

func (cc *ShadowsocksRawConn) Init(parent net.Conn, args interface{}) error {
	if v, ok := args.(ShadowsocksRawConnParams); ok {
		cc.params = v
		cc.Conn = parent

		if cc.params.IsServer {
			cc.forwardReady = make(chan socks.Addr, 1)
		} else {
			//把头部发送出去
			cc.Write(nil)
		}

		return nil
	}

	return fmt.Errorf("arg error:%s", args)
}

func (cc *ShadowsocksRawConn) ForwardReady() <-chan socks.Addr {
	return cc.forwardReady
}

func (cc *ShadowsocksRawConn) Read(b []byte) (n int, err error) {
	if cc.params.IsServer && !cc.isServerTargetRead {
		var tgt socks.Addr

		tgt, err = socks.ReadAddr(cc.Conn)
		if err != nil {
			log.Printf("failed to get target address: %v", err)
			return
		}

		cc.forwardReady <- tgt
		cc.params.Target = tgt
		cc.isServerTargetRead = true
	}

	return cc.Conn.Read(b)
}

func (cc *ShadowsocksRawConn) Write(b []byte) (n int, err error) {
	if !cc.params.IsServer && !cc.isClientTargetWrite {
		if _, err = cc.Conn.Write(cc.params.Target); err != nil {
			log.Printf("failed to send target address: %v", err)
			return
		}
		cc.isClientTargetWrite = true
	}

	if b != nil {
		return cc.Conn.Write(b)
	}

	return
}

func (cc *ShadowsocksRawConn) Close() error {
	return cc.Conn.Close()
}

func (cc *ShadowsocksRawConn) LocalAddr() net.Addr {
	return cc.Conn.LocalAddr()
}

func (cc *ShadowsocksRawConn) RemoteAddr() net.Addr {
	return cc.Conn.RemoteAddr()
}

func (cc *ShadowsocksRawConn) SetDeadline(t time.Time) error {
	return cc.Conn.SetReadDeadline(t)
}

func (cc *ShadowsocksRawConn) SetReadDeadline(t time.Time) error {
	return cc.Conn.SetReadDeadline(t)
}

func (cc *ShadowsocksRawConn) SetWriteDeadline(t time.Time) error {
	return cc.Conn.SetWriteDeadline(t)
}
