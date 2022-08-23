package jsonrpc

import (
	"context"
	"fmt"
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

func NewRelayClient(strUrl string, header http.Header, options ...*RelayOption) *RelayClient {

	return &RelayClient{
		pool: newConnPool(strUrl, header, options...),
	}
}

func newConnPool(strUrl string, header http.Header, options ...*RelayOption) *pool.Pool {
	u, err := url.Parse(strUrl)
	if err != nil {
		log.Panic("parse relay url error [%s]", err.Error())
		return nil
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
			log.Panic("dial ", err)
		}
		return conn
	})
	return p
}

//Call only relay a JSON-RPC request to remote server
func (c *RelayClient) Call(strRequest string) (strResponse string, err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(strRequest))
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

//Send send a JSON-RPC request to remote server and return immediately
func (c *RelayClient) Send(strRequest string) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(strRequest))
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)
	return
}

//Subscribe send a JSON-RPC request to remote server and subscribe this channel (if request is nil, just subscribe)
func (c *RelayClient) Subscribe(ctx context.Context, strRequest string, cb func(c context.Context, msg []byte) bool) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	if strRequest != "" {
		err = conn.WriteMessage(websocket.TextMessage, []byte(strRequest))
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
