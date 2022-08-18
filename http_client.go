package jsonrpc

import (
	"github.com/civet148/httpc"
	"github.com/civet148/log"
	"net/http"
	"sync"
)

type HttpOption struct {
	Timeout int
}

type HttpClient struct {
	strUrl string
	conn *httpc.Client
	requestID int64
	locker sync.RWMutex
}

func NewHttpClient(strUrl string, header http.Header, opts...*HttpOption) (*HttpClient, error) {
	var timeout int
	for _, opt := range opts {
		timeout = opt.Timeout
	}
	return &HttpClient{
		strUrl: strUrl,
		requestID: baseRequestID,
		conn: httpc.NewHttpClient(&httpc.Option{
			Header: header,
			Timeout: timeout,
		}),
	}, nil
}

//POST request to http server
func (c *HttpClient) Call(out interface{}, method string, params...interface{}) (err error){
	request := MakeJSONRpcRequest(genRequestID(), method, params...)
	if err != nil {
		return log.Errorf("make JSON RPC request error [%s]", err.Error())
	}
	var r *httpc.Response
	r, err = c.conn.PostJson(c.strUrl, request)
	if err != nil {
		return log.Errorf("post error [%s]", err.Error())
	}
	var resp JSONRpcResponse
	if err = resp.Unmarshal(r.Body, out); err != nil {
		log.Errorf(err.Error())
		return err
	}
	return nil
}

