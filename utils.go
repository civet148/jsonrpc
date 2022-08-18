package jsonrpc

import "time"

func genRequestID() (id int64) {
	return time.Now().UnixNano()
}
