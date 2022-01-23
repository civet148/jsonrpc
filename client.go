package jsonrpc

import (
	"context"
	"github.com/civet148/log"
	gojsonrpc "github.com/filecoin-project/go-jsonrpc"
	"net/http"
)

type Client struct {
	closer gojsonrpc.ClientCloser
}

func NewClient(ctx context.Context, strUrl, strSpaceName string, requestHeader http.Header, handlers ...interface{}) (*Client, error) {
	closer, err := gojsonrpc.NewMergeClient(ctx, strUrl, strSpaceName, handlers, requestHeader)
	if err != nil {
		log.Errorf("new json rpc client error [%s]", err.Error())
		return nil, err
	}
	return &Client{
		closer: closer,
	}, nil
}

func (m *Client) Close() {
	m.closer()
}
