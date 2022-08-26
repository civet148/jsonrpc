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

type RelayOption struct {
}

type RelayClient struct {
	pool   *pool.Pool
	closed bool
}

func NewRelayClient(strUrl string, header http.Header, options ...*RelayOption) (*RelayClient, error) {
	p, err := newConnPool(strUrl, header, options...)
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	return &RelayClient{
		pool: p,
	}, nil
}

func newConnPool(strUrl string, header http.Header, options ...*RelayOption) (*pool.Pool, error) {
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
	var p = pool.New(func() interface{} {
		var conn *websocket.Conn
		conn, _, err = websocket.DefaultDialer.Dial(u.String(), header)
		if err != nil {
			log.Errorf(err.Error())
			return nil
		}
		return conn
	})
	return p, nil
}

func (c *RelayClient) getConn() (*websocket.Conn, error) {
	var conn *websocket.Conn
	ws := c.pool.Get()
	if ws == nil {
		return nil, log.Errorf("nil websocket connection")
	}
	conn = ws.(*websocket.Conn)
	if conn == nil {
		return nil, log.Errorf("websocket connection is nil")
	}
	return conn, nil
}

//Call only relay a JSON-RPC request to remote server
func (c *RelayClient) Call(request []byte) (strResponse string, err error) {
	var conn *websocket.Conn
	conn, err = c.getConn()
	if err != nil {
		return "", log.Errorf(err.Error())
	}
	err = conn.WriteMessage(websocket.TextMessage, request)
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

//CallNoReply send a JSON-RPC request to remote server and return immediately
func (c *RelayClient) CallNoReply(request []byte) (err error) {
	var conn *websocket.Conn
	conn, err = c.getConn()
	if err != nil {
		return log.Errorf(err.Error())
	}
	err = conn.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)
	return
}

//Subscribe send a JSON-RPC request to remote server and subscribe this channel (if request is nil, just subscribe)
func (c *RelayClient) Subscribe(ctx context.Context, request []byte, cb func(c context.Context, msg []byte) bool) (err error) {
	var conn *websocket.Conn
	conn, err = c.getConn()
	if err != nil {
		return log.Errorf(err.Error())
	}
	if len(request) != 0 {
		err = conn.WriteMessage(websocket.TextMessage, request)
		if err != nil {
			log.Errorf("write message error [%s]", err.Error())
			_ = conn.Close() //broken pipe maybe
			return
		}
	}

	for {
		if c.closed {
			_ = conn.Close()
			break
		}
		var msg []byte
		_, msg, err = conn.ReadMessage()
		if err != nil {
			log.Errorf("read message error [%s]", err.Error())
			break
		}
		if ok := cb(ctx, msg); ok == false {
			break //stop subscribe
		}
	}
	if !c.closed {
		c.pool.Put(conn)
	}
	return
}

func (c *RelayClient) Close() {
	c.pool.RemoveAll()
	c.closed = true
}
