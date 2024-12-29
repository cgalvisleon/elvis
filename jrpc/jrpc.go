package jrpc

import (
	"net/rpc"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

func RpcCall(host string, port int, method string, data et.Json) (et.Item, error) {
	var args []byte = data.ToByte()
	var reply *[]byte

	client, err := rpc.DialHTTP("tcp", strs.Format(`%s:%d`, host, port))
	if err != nil {
		return et.Item{}, console.Error(err)
	}
	defer client.Close()

	err = client.Call(method, args, &reply)
	if err != nil {
		return et.Item{}, console.Error(err)
	}

	result := et.Json{}.ToItem(*reply)

	return result, nil
}
