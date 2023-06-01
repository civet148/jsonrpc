package jsonrpc

import (
	"context"
	"github.com/civet148/log"
	"github.com/civet148/pool"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strings"
)

type SubFunc func(context.Context, []byte) bool

type RelayOption struct {
}

type SubNotify struct {
	cb  SubFunc
	ctx context.Context
}

type RelayClient struct {
	sub     *RelayConn     //subscribe connection
	pool    *pool.Pool     //rpc call connection pool
	channel chan SubNotify //rpc subscribe channel
}

func NewRelayClient(strUrl string, header http.Header, options ...*RelayOption) (*RelayClient, error) {
	p, err := newConnPool(strUrl, header, options...)
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	sub, err := newConn(strUrl, header, options...)
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	c := &RelayClient{
		pool:    p,
		sub:     sub,
		channel: make(chan SubNotify),
	}
	go c.selectSubChannel()
	return c, nil
}

func newConnPool(strUrl string, header http.Header, options ...*RelayOption) (*pool.Pool, error) {
	var p = pool.New(func() interface{} {
		var err error
		var conn *RelayConn
		conn, err = newConn(strUrl, header, options...)
		if err != nil {
			return log.Errorf(err.Error())
		}
		return conn
	})
	return p, nil
}

func newConn(strUrl string, header http.Header, options ...*RelayOption) (conn *RelayConn, err error) {
	u, err := parseUrl(strUrl)
	if err != nil {
		log.Panic("parse relay url error [%s]", err.Error())
		return nil, err
	}
	var wc *websocket.Conn
	wc, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	return &RelayConn{
		conn: wc,
	}, nil
}

func parseUrl(strUrl string) (*url.URL, error) {
	u, err := url.Parse(strUrl)
	if err != nil {
		log.Panic("parse relay url error [%s]", err.Error())
		return nil, err
	}
	if u.Scheme != UrlSchemeWS && u.Scheme != UrlSchemeWSS {
		if "http" == strings.ToLower(u.Scheme) {
			u.Scheme = "ws"
		}
		if "https" == strings.ToLower(u.Scheme) {
			u.Scheme = "wss"
		}
	}
	return u, nil
}

func (c *RelayClient) getConn() (*RelayConn, error) {
	var conn *RelayConn
	ws := c.pool.Get()
	if ws == nil {
		return nil, log.Errorf("nil websocket connection")
	}
	conn = ws.(*RelayConn)
	if conn == nil {
		return nil, log.Errorf("websocket connection is nil")
	}
	return conn, nil
}

//Call only relay a JSON-RPC request to remote server
func (c *RelayClient) Call(request []byte) (strResponse string, err error) {
	var conn *RelayConn
	conn, err = c.getConn()
	if err != nil {
		return "", log.Errorf(err.Error())
	}
	err = conn.WriteMessage(websocket.BinaryMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)

	var msg []byte
	_, msg, err = conn.ReadMessage()
	strResponse = string(msg)
	return
}

func (c *RelayClient) selectSubChannel() {
	var running bool
	for {
		select {
		case cb := <-c.channel:
			{
				if !running {
					running = true
					go c.readFromSubscribeSocket(cb)
				}
			}
		}
	}

}

func (c *RelayClient) readFromSubscribeSocket(sn SubNotify) {
	var err error
	conn := c.sub
	log.Debugf("subscribe reading...")
	for {
		var msg []byte
		_, msg, err = conn.ReadMessage()
		if err != nil {
			log.Warnf("read websocket error [%s] socket closed", err.Error())
			break
		}
		if ok := sn.cb(sn.ctx, msg); ok == false {
			log.Warnf("callback return false, subscribe will stop")
			break //subscribe stopped
		}
	}
}

//CallNoReply send a JSON-RPC request to remote server and return immediately
func (c *RelayClient) CallNoReply(request []byte) (err error) {
	var conn *RelayConn
	conn, err = c.getConn()
	if err != nil {
		return log.Errorf(err.Error())
	}
	err = conn.WriteMessage(websocket.BinaryMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)
	return
}

//Subscribe send a JSON-RPC request to remote server and subscribe this channel (if request is nil, just subscribe)
func (c *RelayClient) Subscribe(ctx context.Context, request []byte, cb func(context.Context, []byte) bool) (err error) {
	var conn *RelayConn
	conn = c.sub
	if len(request) == 0 || cb == nil {
		return log.Errorf("empty request or nil callback")
	}
	err = conn.WriteMessage(websocket.BinaryMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	c.channel <- SubNotify{
		cb:  cb,
		ctx: ctx,
	}
	return
}

func (c *RelayClient) Close() {
	c.pool.RemoveAll()
	if c.sub != nil {
		c.sub.Close()
	}
}
