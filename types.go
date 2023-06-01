package jsonrpc

import (
	"encoding/json"
	"github.com/civet148/log"
	"github.com/gorilla/websocket"
	"sync"
)

const (
	baseRequestID = 10000
)

const (
	UrlSchemeWS  = "ws"
	UrlSchemeWSS = "wss"
)

type JSONRpcRequest struct {
	Id      int64       `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JSONRpcResponse struct {
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Error   *JSONRpcError `json:"error"`
	Result  interface{}   `json:"result"`
}

type JSONRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *JSONRpcRequest) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *JSONRpcResponse) Unmarshal(data []byte, v interface{}) error {
	m.Result = v
	err := json.Unmarshal(data, m)
	if err != nil {
		return log.Errorf("unmarshal response data error [%s]", err.Error())
	}
	if m.Error != nil {
		return log.Errorf("JSON RPC error code [%v] message [%s]", m.Error.Code, m.Error.Message)
	}
	return nil
}

func MakeJSONRpcRequest(id int64, method string, params ...interface{}) *JSONRpcRequest {
	var req *JSONRpcRequest
	req = &JSONRpcRequest{
		Id:      id,
		Jsonrpc: "2.0",
		Method:  method,
	}
	n := len(params)
	if n == 1 {
		req.Params = params[0]
	} else if n >= 2 {
		req.Params = params
	}
	return req
}

func MarshalJSONRpcRequest(id int64, method string, params ...interface{}) (data []byte, err error) {
	req := MakeJSONRpcRequest(id, method, params...)
	return req.Marshal()
}

type RelayConn struct {
	locker sync.Mutex //rpc write lock
	conn   *websocket.Conn
}

func (c *RelayConn) Close() error {
	return c.conn.Close()
}

func (c *RelayConn) ReadMessage() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *RelayConn) WriteMessage(msgType int, data []byte) error {
	c.locker.Lock()
	defer c.locker.Unlock()
	return c.conn.WriteMessage(msgType, data)
}
