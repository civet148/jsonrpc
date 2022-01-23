package main

import (
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
	"net/http"
)

/*
	test JSON-RPC datagram relay to other JSON-RPC server
*/

var strRequest = "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"Filecoin.Version\",\"params\":[],\"meta\":{\"SpanContext\":\"AAC7j+ekzez0hIctfuIn9xJCASdQ/uZFvPaeAgA=\"}}"

func main() {
	var strUrl = "ws://192.168.20.108:1234/rpc/v0"
	var strToken = "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.sG1ess9q1Q5B_QMKoiq0IAR9YsnYNgKWoC7-NphGnhM"
	header := http.Header{}
	header.Add("Authorization", "Bearer "+strToken)
	relay := jsonrpc.NewRelayClient(strUrl, header)
	defer relay.Close()
	strResponse, err := relay.Call(strRequest)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("RPC relay response [%s]", strResponse)
}
