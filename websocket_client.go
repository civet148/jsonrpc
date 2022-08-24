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
	"sync"
)


type WebSocketClient struct {
	pool   *pool.Pool
	closed bool
	requestID int64
	locker sync.RWMutex
}

func NewWebSocketClient(strUrl string, header http.Header) *WebSocketClient {

	return &WebSocketClient{
		requestID: baseRequestID,
		pool: newWebSocketPool(strUrl, header),
	}
}

func newWebSocketPool(strUrl string, header http.Header) *pool.Pool {
	u, err := url.Parse(strUrl)
	if err != nil {
		log.Panic("parse relay url error [%s]", err.Error())
		return nil
	}
	log.Infof("header [%+v]", header)
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

//Call submit a JSON-RPC request to remote server
func (c *WebSocketClient) Call(out interface{}, method string, params ...interface{}) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		return log.Errorf("websocket connection pool is nil")
	}
	var request []byte
	request, err = MarshalJSONRpcRequest(genRequestID(), method, params...)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Debugf("request [%s]", request)
	err = conn.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)

	var msg []byte
	_, msg, err = conn.ReadMessage()
	var resp JSONRpcResponse
	log.Debugf("response [%s]", msg)
	if err = resp.Unmarshal(msg, out); err != nil {
		return err
	}
	return nil
}

//SubscribeCall send a JSON-RPC request to remote server and subscribe this channel (if method is nil, just subscribe)
func (c *WebSocketClient) SubscribeCall(ctx context.Context, cb func(ctx context.Context, msg []byte) bool, method string, params ...interface{}) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	if method != "" {
		var request []byte
		request, err = MarshalJSONRpcRequest(genRequestID(), method, params...)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
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


//Subscribe send a request to remote server and subscribe this channel (if request is nil, just subscribe)
func (c *WebSocketClient) Subscribe(ctx context.Context, request []byte, cb func(c context.Context, msg []byte) bool) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
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

//CallNoReply send a JSON-RPC request to remote server and return immediately
func (c *WebSocketClient) CallNoReply(method string, params ...interface{}) (err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	if method != "" {
		var request []byte
		request, err = MarshalJSONRpcRequest(genRequestID(), method, params...)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
		err = conn.WriteMessage(websocket.TextMessage, request)
		if err != nil {
			log.Errorf("write message error [%s]", err.Error())
			_ = conn.Close() //broken pipe maybe
			return
		}
	}
	defer c.pool.Put(conn)
	return
}


//Send send a request to remote server and return immediately
func (c *WebSocketClient) Send(request []byte) (err error) {
	if len(request) == 0 {
		return log.Errorf("request is empty")
	}
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	defer c.pool.Put(conn)
	err = conn.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}

	return
}

func (c *WebSocketClient) Close() {
	c.pool.RemoveAll()
	c.closed = true
}
