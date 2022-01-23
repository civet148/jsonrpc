package jsonrpc

import (
	"github.com/civet148/log"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

type RelayClient struct {
	conn *websocket.Conn
}

func NewRelayClient(strUrl string, header http.Header) *RelayClient {
	u, err := url.Parse(strUrl)
	if err != nil {
		log.Panic("parse relay url error [%s]", err.Error())
		return nil
	}
	if u.Scheme != UrlSchemeWS && u.Scheme != UrlSchemeWSS {
		log.Panic("url scheme not 'ws' or 'wss'")
		return nil
	}
	var conn *websocket.Conn
	conn, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Panic("dial ", err)
	}
	return &RelayClient{
		conn: conn,
	}
}

//Call only relay a JSON-RPC request json string to remote server
func (c *RelayClient) Call(strRequest string) (strResponse string, err error) {
	err = c.conn.WriteMessage(websocket.TextMessage, []byte(strRequest))
	if err != nil {
		log.Errorf("WriteMessage error [%s]", err.Error())
		return
	}
	return
}
