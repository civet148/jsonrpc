package main

import (
	"context"
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
	"net/http"
)

/*
	test JSON-RPC datagram relay to other JSON-RPC server
*/

var strRequest = "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"Filecoin.Version\",\"params\":[],\"meta\":{\"SpanContext\":\"AAC7j+ekzez0hIctfuIn9xJCASdQ/uZFvPaeAgA=\"}}"

func main() {
	//var strUrl = "ws://192.168.20.108:1234/rpc/v0"
	//var strToken = "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.sG1ess9q1Q5B_QMKoiq0IAR9YsnYNgKWoC7-NphGnhM"
	header := http.Header{}
	//header.Add("Authorization", "Bearer "+strToken)
	//relay := jsonrpc.NewRelayClient(strUrl, header)
	//defer relay.Close()
	//strResponse, err := relay.Call(strRequest)
	//if err != nil {
	//	log.Errorf(err.Error())
	//	return
	//}
	//log.Infof("RPC relay response [%s]", strResponse)
	//relay.Close()

	strRequest = `{
  "topic":"9d9c764d-e6ca-4163-9320-779402e880d3",
  "type":"pub",
  "payload":"{\"data\":\"42eb083fa4e59bc0e184aa3d938af7a01bbc943345f96bc4d17d39b267312a9362e27baa43a488d6ed08d7e29a6fe4bf63e4a1c19e06bfedee142f3e9bf6d1308871efeea93a8044814fb02c3581b80933b290ab730d2da978aee35d36ee2f9153736413b875ec9f8255245f33dbf8aed11e3c710d3b3de5c0445078c60377745fb56208838e43cb8564108c21380bb6d9a6beb1a97172a562dd07d2d0eeaeae2320115932518c63be41578d15c76e0ce2d53ad43974a00512a6e8c9e280e51a38da98534c034a31e789a2acc2bd6656a4972da81d8dcea83254160e39a46f39753fa450fc6cd68d0086116c8cb22a360cbe3598528b686865c59e33bd918301bcb6040959bac57fc612f6428d9e36a95cbc11223b055da7aa94571537f424ae44035bbd70cd6f8310f134b73e258ccc20716432589bd90e03a3818d8a5d9159ed7647c574180689c80b4639ab598ebab5fb2f2798f391f5080fa1493595ff936264b028c6c06a3c37c7e5f3135974b36a2d3fe85acdf497976a1856ece47f2578a3d0772faf6fcd3d47a20803f1d100d41697ba62927bfc58efa67e369530acde5a75fd53765d30e406fc4ca5557fcd2833835c8a2e0c49d2a9124b2f1bf13b\",\"hmac\":\"c8bf96199abe0059fd1a8869448516d871e56201e6759f71c03f39be2e2bbd2a\",\"iv\":\"0b0026d95c2b6704b9e02c9bd34dcfa0\"}",
  "silent":true
}`
	log.Infof("sending request...")
	relay, err := jsonrpc.NewRelayClient("wss://0.bridge.walletconnect.org/?env=browser&host=app.uniswap.org&protocol=wc&version=1", header)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer relay.Close()
	err = relay.CallNoReply([]byte(strRequest))
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	strRequest = `{
  "topic":"28f17b68-4878-4473-8bc3-15f6af81a773",
  "type":"sub",
  "payload":"",
  "silent":true
}`
	if err = relay.Subscribe(context.TODO(), []byte(strRequest), sub); err != nil {
		log.Errorf(err.Error())
		return
	}
}

func sub(ctx context.Context, msg []byte) bool {
	log.Infof("sub data [%s]", msg)
	return false
}