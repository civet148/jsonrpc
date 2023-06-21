package main

import (
	"context"
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
	"time"
)

/*
	test JSON-RPC datagram relay to other JSON-RPC server
*/

func main() {
	var strUrl = "ws://172.16.35.31:1234/rpc/v0"
	relay, err := jsonrpc.NewRelayClient(strUrl, nil)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer relay.Close()
	req1 := "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"Filecoin.ChainNotify\",\"params\":[],\"meta\":{\"SpanContext\":\"AAB/fIpdCiYJp3/GcvLY8HbLAfx3mBcbkC+UAgA=\"}}"
	go func() {
		err = subscribe(relay, req1, true)
		if err != nil {
			log.Errorf("sub 1 error [%s]", err.Error())
		}
	}()

	time.Sleep(3 * time.Second) //wait for subscribe routine 1 startup

	req2 := "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"Filecoin.ChainNotify\",\"params\":[],\"meta\":{\"SpanContext\":\"AAA6xgjut6cbxaJQHN1tqNjDAUHQBQSFdtmdAgA=\"}}"
	go func() {
		err = subscribe(relay, req2, true)
		if err != nil {
			log.Errorf("sub 2 error [%s]", err.Error())
		}
	}()

	time.Sleep(60 * time.Minute)
}

func subscribe(relay *jsonrpc.RelayClient, req string, block bool) (err error) {
	var opt = &jsonrpc.SubscribeOption{Block: block}
	if err = relay.Subscribe(context.Background(), []byte(req), callback, opt); err != nil {
		return log.Errorf(err.Error())
	}
	return nil
}

func callback(ctx context.Context, msg []byte) bool {
	log.Infof("sub data [%s]", msg)
	return true
}
