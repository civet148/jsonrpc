package jsonrpc

import (
	"fmt"
	"github.com/civet148/log"
	"github.com/civet148/pool"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

type RelayOption struct {
}

type RelayClient struct {
	pool *pool.Pool
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
		log.Panic("url scheme not 'ws' or 'wss'")
		return nil
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

//Call only relay a JSON-RPC request json string to remote server
func (c *RelayClient) Call(strRequest string) (strResponse string, err error) {
	var conn *websocket.Conn
	conn = c.pool.Get().(*websocket.Conn)
	if conn == nil {
		err = fmt.Errorf("websocket connection is nil")
		log.Errorf(err.Error())
		return
	}
	defer c.pool.Put(conn)
	err = conn.WriteMessage(websocket.TextMessage, []byte(strRequest))
	if err != nil {
		log.Errorf("WriteMessage error [%s]", err.Error())
		return
	}
	var msg []byte
	_, msg, err = conn.ReadMessage()
	strResponse = string(msg)
	return
}

func (c *RelayClient) Close() {

}
