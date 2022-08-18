package main

import (
	"context"
	"fmt"
	"github.com/civet148/jsonrpc"
	"github.com/civet148/jsonrpc/examples/common"
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
	var strRemoteAddr = common.GatewayUrl
	c, err := jsonrpc.NewMergeClient(context.Background(), strRemoteAddr, common.GatewayNamespace, nil, &client)
	if err != nil {
		fmt.Printf("jsonrpc.NewClient error [%s]\n", err.Error())
		return
	}
	defer c.Close()

	log.Infof("namespace [%s] connect address [%s] ok\n", common.GatewayNamespace, strRemoteAddr)
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
