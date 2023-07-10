# jsonrpc
json rpc server and client

# server
```golang
package main

import (
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
)

type GatewayAPI interface {
	Add(a, b int) (int, error)
	Sub(a, b int) (int, error)
	Mul(a, b int) (int, error)
	Div(a, b int) (int, error)
	Ping() (string, error)
}

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
	var strNamespace = "Gateway"
	serverHandler := &GatewayServer{}
	//listen http://host:port/rpc/v0 for RPC call with internal HTTP server
	var strUrl = "ws://127.0.0.1:8000/rpc/v0"
	rpcServer := jsonrpc.NewServer(strNamespace, serverHandler)
	log.Infof("namespace [%s] url [%s] listening...", strNamespace, strUrl)
	if err := rpcServer.ListenAndServe(strUrl); err != nil {
		log.Panic(err.Error())
		return
	}
}
```

# client
```golang

import (
	"context"
	"fmt"
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
	"runtime"
)

type GatewayClient struct {
	Add  func(a, b int) (int, error)
	Sub  func(a, b int) (int, error)
	Mul  func(a, b int) (int, error)
	Div  func(a, b int) (int, error)
	Ping func() (string, error)
}

func init() {
	switch runtime.GOARCH {
	case "amd64":
	default:
		panic("GOARCH is not amd64")
	}
}

func main() {
	var client GatewayClient
    var strNamespace = "Gateway"
	var strRemoteAddr = "ws://127.0.0.1:8000/rpc/v0"
	c, err := jsonrpc.NewClient(context.Background(), strRemoteAddr, strNamespace, nil, &client)
	if err != nil {
		fmt.Printf("jsonrpc.NewClient error [%s]\n", err.Error())
		return
	}
	defer c.Close()

	log.Infof("namespace [%s] connect address [%s] ok\n", strNamespace, strRemoteAddr)
	var n int
	var pong string
	pong, err = client.Ping() //expect 'pong' return from server
	if err != nil {
		log.Errorf("client.Ping error [%s]\n", err.Error())
		return
	}
	log.Infof("client.Ping result [%s]\n", pong)

	n, err = client.Add(1, 2) //expect 3 return from server
	if err != nil {
		log.Errorf("client.Add error [%s]\n", err.Error())
		return
	}
	log.Infof("client.Add result [%d]\n", n)

	n, err = client.Sub(2, 1) //expect 1 return from server
	if err != nil {
		log.Errorf("client.Sub error [%s]\n", err.Error())
		return
	}
	log.Infof("client.Sub result [%d]\n", n)

	n, err = client.Mul(1, 2) //expect 2 return from server
	if err != nil {
		log.Errorf("client.Mul error [%s]\n", err.Error())
		return
	}
	log.Infof("client.Mul result [%d]\n", n)

	n, err = client.Div(2, 1) //expect 2 return from server
	if err != nil {
		log.Errorf("client.Div error [%s]\n", err.Error())
		return
	}
	log.Infof("client.Div result [%d]\n", n)
}
```