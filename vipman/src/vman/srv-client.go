package vman

import (
	"net/rpc"
)

func RpcListCall() {
}

func RpcStatusCall() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:12345")
	if err != nil {
		Panic("dialing: %v\n", err)
		return
	}

	args := &EchoArgs{"ping"}
	var reply string
	err = client.Call(ECHO_M, args, &reply)
	if err != nil {
		Panic("echo error: %v\n", err)
	}

	Log("RPC: %s -> %s \n", args.Request, reply)
}
