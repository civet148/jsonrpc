package main

import (
	"github.com/civet148/jsonrpc"
	"github.com/civet148/jsonrpc/examples/api"
	"github.com/civet148/jsonrpc/examples/common"
	"github.com/civet148/log"
)

type GatewayServer struct {
}

//ensure gateway server implemented api.GatewayAPI
var _ api.GatewayAPI = (*GatewayServer)(nil)

func (m *GatewayServer) Add(a, b int) (int, error) {
	return a + b, nil
}

func (m *GatewayServer) Sub(a, b int) (int, error) {
	return a - b, nil
}

func (m *GatewayServer) Mul(a, b int) (int, error) {
	return a * b, nil
}

func (m *GatewayServer) Div(a, b int) (int, error) {
	return a / b, nil
}

func (m *GatewayServer) Ping() (string, error) {
	return "Pong", nil
}

func main() {
	TestRPCServer()
}

func TestRPCServer() {

	serverHandler := &GatewayServer{}
	////listen http://host:port/ for RPC call with standard HTTP server
	//rpcServer := jsonrpc.NewServer(common.GatewayNamespace, serverHandler)
	//log.Infof("namespace [%s] address [%s] listening...\n", common.GatewayNamespace, common.GatewayHttpAddr)
	//if err := http.ListenAndServe(common.GatewayHttpAddr, rpcServer); err != nil {
	//	log.Errorf(err.Error())
	//	return
	//}

	//listen http://host:port/rpc/v0 for RPC call with internal HTTP server
	var strUrl = common.GatewayUrl
	rpcServer := jsonrpc.NewServer(common.GatewayNamespace, serverHandler)

	log.Infof("namespace [%s] url [%s] listening...", common.GatewayNamespace, strUrl)
	if err := rpcServer.ListenAndServe(strUrl); err != nil {
		log.Panic(err.Error())
		return
	}
}
