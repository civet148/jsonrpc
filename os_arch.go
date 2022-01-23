package jsonrpc

import "runtime"

func init() {
	switch runtime.GOARCH {
	case "amd64":
	default:
		panic("GOARCH is not amd64")
	}
}
