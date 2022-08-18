package jsonrpc

import (
	"context"
	"github.com/civet148/log"
	gojsonrpc "github.com/filecoin-project/go-jsonrpc"
	"net/http"
)

type MergeClient struct {
	closer gojsonrpc.ClientCloser
}

func NewMergeClient(ctx context.Context, strUrl, strSpaceName string, requestHeader http.Header, handlers ...interface{}) (*MergeClient, error) {
	closer, err := gojsonrpc.NewMergeClient(ctx, strUrl, strSpaceName, handlers, requestHeader)
	if err != nil {
		log.Errorf("new json rpc client error [%s]", err.Error())
		return nil, err
	}
	return &MergeClient{
		closer: closer,
	}, nil
}

func (m *MergeClient) Close() {
	m.closer()
}
