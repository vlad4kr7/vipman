package vman

import (
	"fmt"
	"log"
	"net/rpc"
)

func RpcListCall() {
}

func RpcStatusCall() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal("dialing:", err)
		return
	}

	args := &EchoArgs{"ping"}
	var reply string
	err = client.Call(ECHO_M, args, &reply)
	if err != nil {
		log.Fatal("echo error:", err)
	}

	fmt.Printf("RPC: %s -> %s \n", args.Request, reply)
}
