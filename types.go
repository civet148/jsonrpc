package jsonrpc

import "encoding/json"

const(
	baseRequestID = 10000
)

const (
	UrlSchemeWS  = "ws"
	UrlSchemeWSS = "wss"
)

type JSONRpcRequest struct {
	Id      int64    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []interface{} `json:"params"`
}

type JSONRpcResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}


func (m *JSONRpcRequest) Marshal() ([]byte, error){
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *JSONRpcResponse) Unmarshal(data []byte, v interface{}) error {
	m.Result = v
	return json.Unmarshal(data, m)
}

func MakeJSONRpcRequest(id int64, method string, params... interface{}) *JSONRpcRequest {
	req := &JSONRpcRequest{
		Id:      id,
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}
	return req
}

func MarshalJSONRpcRequest(id int64, method string, params... interface{}) (data []byte, err error) {
	req := MakeJSONRpcRequest(id, method, params...)
	return req.Marshal()
}

