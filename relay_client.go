package jsonrpc

import (
	"context"
	"github.com/civet148/log"
	"github.com/civet148/pool"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	allowMaxQPS = 10000 //only max qps allowed
)

type SubFunc func(context.Context, []byte) bool

type RelayOption struct {
	MaxQPS         int //rate limit of QPS (0 means no limit)
	MaxConnections int //maximum number of connections (0 means no limit)
	MsgType        int //websocket message type
}

type SubscribeOption struct {
	Block bool //blocking subscribe channel
}

type SubNotify struct {
	cb  SubFunc
	ctx context.Context
}

type RelayClient struct {
	subscribed bool           //is subscribed
	locker     sync.RWMutex   //internal lock
	sub        *RelayConn     //subscribe connection
	pool       *pool.Pool     //rpc call connection pool
	ready      chan bool      //is all works ready
	notify     chan SubNotify //rpc subscribe channel
	cbs        []SubFunc      //rpc callback functions
	opt        *RelayOption   //relay option
	limiter    *rate.Limiter  //rate limit
	count      int            //connection count
	strUrl     string         //relay url
}

func NewRelayClient(strUrl string, header http.Header, options ...*RelayOption) (*RelayClient, error) {
	var opt *RelayOption
	if len(options) != 0 {
		opt = options[0]
		if opt.MsgType == 0 {
			opt.MsgType = websocket.TextMessage
		}
	} else {
		opt = makeDefaultRelayOption()
	}

	var err error
	var limiter *rate.Limiter
	if opt.MaxQPS > 0 {
		qps := opt.MaxQPS
		if qps > allowMaxQPS {
			qps = allowMaxQPS
			log.Warnf("max qps is limited to %d", qps)
		}
		limiter = rate.NewLimiter(rate.Limit(qps), qps)
	}
	c := &RelayClient{
		opt:     opt,
		limiter: limiter,
		strUrl:  strUrl,
		ready:   make(chan bool),
		notify:  make(chan SubNotify),
	}
	c.pool, err = c.newConnPool(strUrl, header)
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	c.sub, err = c.newConn(strUrl, header)
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	go c.selectSubChannel()
	return c, nil
}

func makeDefaultRelayOption() *RelayOption {
	return &RelayOption{
		MsgType: websocket.TextMessage,
	}
}

func (c *RelayClient) newConnPool(strUrl string, header http.Header) (*pool.Pool, error) {
	var p = pool.New(func() interface{} {
		var err error
		var conn *RelayConn
		if c.opt != nil && c.opt.MaxConnections > 0 {
			if c.count > c.opt.MaxConnections {
				log.Warnf("connection count is out of limit %d", c.opt.MaxConnections)
				return nil
			}
		}
		conn, err = c.newConn(strUrl, header)
		if err != nil {
			log.Errorf("new connection error [%s]", err.Error())
			return nil
		}
		c.count++
		return conn
	})
	return p, nil
}

func (c *RelayClient) GetConnCount() int {
	return c.count
}

func (c *RelayClient) newConn(strUrl string, header http.Header) (conn *RelayConn, err error) {
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
	return conn, nil
}

func (c *RelayClient) limitWait() error {
	if c.limiter != nil {
		err := c.limiter.Wait(context.Background())
		if err != nil {
			return log.Errorf("rate limit wait error [%s]", err)
		}
	}
	return nil
}

// Call only relay a JSON-RPC request to remote server
func (c *RelayClient) Call(request []byte) (response []byte, err error) {
	var conn *RelayConn
	err = c.limitWait()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	conn, err = c.getConn()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	err = conn.WriteMessage(c.opt.MsgType, request)
	if err != nil {
		_ = conn.Close() //broken pipe maybe
		return nil, log.Errorf("write message error [%s]", err.Error())
	}
	defer c.pool.Put(conn)
	_, response, err = conn.ReadMessage()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	return
}

func (c *RelayClient) selectSubChannel() {
	var running bool
	for {
		select {
		case cb := <-c.notify:
			{
				if !running {
					running = true
					go c.readSubSocket(cb)
				}
				c.ready <- true
			}
		}
	}
}

// CallNoReply send a JSON-RPC request to remote server and return immediately
func (c *RelayClient) CallNoReply(request []byte) (err error) {
	var conn *RelayConn
	conn, err = c.getConn()
	if err != nil {
		return log.Errorf(err.Error())
	}
	err = conn.WriteMessage(c.opt.MsgType, request)
	if err != nil {
		log.Errorf("write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	defer c.pool.Put(conn)
	return
}

// Subscribe send a JSON-RPC request to remote server and subscribe this channel (if request is nil, just subscribe)
func (c *RelayClient) Subscribe(ctx context.Context, request []byte, cb func(context.Context, []byte) bool, options ...*SubscribeOption) (err error) {
	if cb == nil {
		return log.Errorf("empty callback")
	}

	sn := SubNotify{
		cb:  cb,
		ctx: ctx,
	}
	var block bool
	if len(options) != 0 {
		opt := options[0]
		if opt.Block {
			block = true
		}
	}

	if block {
		c.locker.Lock()
		if c.subscribed {
			c.locker.Unlock()
			return log.Errorf("this channel is already subscribed")
		}
		c.subscribed = true
		c.locker.Unlock()
		if err = c.writeSubSocket(request); err != nil {
			return log.Errorf(err.Error())
		}
		c.readSubSocket(sn)
	} else {
		c.locker.Lock()
		defer c.locker.Unlock()
		if err = c.writeSubSocket(request); err != nil {
			return log.Errorf(err.Error())
		}
		c.notify <- sn
		select {
		case _ = <-c.ready:
			{
				c.subscribed = true
			}
		}
	}
	return
}

func (c *RelayClient) Close() {
	c.pool.RemoveAll()
	if c.sub != nil {
		c.sub.Close()
	}
}

func (c *RelayClient) readSubSocket(sn SubNotify) {
	var err error
	for {
		var msg []byte
		_, msg, err = c.sub.ReadMessage()
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

func (c *RelayClient) writeSubSocket(msg []byte) (err error) {
	conn := c.sub
	err = c.sub.WriteMessage(c.opt.MsgType, msg)
	if err != nil {
		log.Errorf("subscribe websocket write message error [%s]", err.Error())
		_ = conn.Close() //broken pipe maybe
		return
	}
	return nil
}
