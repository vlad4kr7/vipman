package vman

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)


type EchoArgs struct {
	Request string
}

type Echo string

const ECHO_M = "Echo.Handler"

func (t *Echo) Handle(args *EchoArgs, reply *string) error {
	*reply = "reply: "+args.Request
	return nil
}

var startedRpc = false

func RpcInit(){
	rpc.Register(new(Echo))
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "127.0.0.1:12345")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	startedRpc = true
	log.Print("listen....\n")
	// go background
 	http.Serve(l, nil)
}
