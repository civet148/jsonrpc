package jsonrpc

import (
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"time"
)

type Option struct {
	MaxRequestSize int64 //max request body size, default 100MiB
	PingInterval   int   //server ping interval, default 5 seconds
}

type MergeServer struct {
	rpcServer *jsonrpc.RPCServer
}

func NewMergeServer(strNamespace string, handler interface{}, options ...*Option) *MergeServer {
	var so *Option
	var opts []jsonrpc.ServerOption
	for _, so = range options {
		if so.MaxRequestSize > 0 {
			opts = append(opts, jsonrpc.WithMaxRequestSize(so.MaxRequestSize))
		}
		if so.PingInterval > 0 {
			opts = append(opts, jsonrpc.WithServerPingInterval(time.Duration(so.PingInterval)*time.Second))
		}
	}
	rpcServer := jsonrpc.NewServer(opts...)
	rpcServer.Register(strNamespace, handler)
	return &MergeServer{
		rpcServer: rpcServer,
	}
}

// ServeHTTP http handler
func (m *MergeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.rpcServer.ServeHTTP(w, r)
}

// AliasMethod alias method to new name
func (m *MergeServer) AliasMethod(alias, original string) {
	m.rpcServer.AliasMethod(alias, original)
}

// ListenAndServe start a http server (NOTE: routine will be blocked)
func (m *MergeServer) ListenAndServe(strUrl string) (err error) {
	u, err := url.Parse(strUrl)
	if err != nil {
		err = fmt.Errorf("parse url [%s] error [%s]", strUrl, err)
		panic(err.Error())
	}
	router := mux.NewRouter()
	router.Handle(u.Path, m.rpcServer)
	if err = http.ListenAndServe(u.Host, router); err != nil {
		return err
	}
	return nil
}
