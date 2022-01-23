package jsonrpc

import (
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

type Server struct {
	rpcServer *jsonrpc.RPCServer
}

func NewServer(strNamespace string, handler interface{}) *Server {
	rpcServer := jsonrpc.NewServer()
	rpcServer.Register(strNamespace, handler)
	return &Server{
		rpcServer: rpcServer,
	}
}

//ServeHTTP http handler
func (m *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.rpcServer.ServeHTTP(w, r)
}

//AliasMethod alias method to new name
func (m *Server) AliasMethod(alias, original string) {
	m.rpcServer.AliasMethod(alias, original)
}

//ListenAndServe start a http server (NOTE: routine will be blocked)
func (m *Server) ListenAndServe(strUrl string) (err error) {
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
