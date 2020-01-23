/*
Author: Yasuhiro Matsumoto (a.k.a mattn)
Origin: https://github.com/mattn/VipmanRPC
Licence: MIT
Changes: alot
Changed by Vlad Krinitsyn
*/
package vman

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// VipmanRPC is RPC server
type VipmanRPC struct{} // rpcChan chan<- *rpcMessage

type rpcMessage struct {
	Msg  string
	Args []string
	// sending error (if any) when the task completes
	ErrCh chan error
}

// TODO List of child statuses
func (r *VipmanRPC) List(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	for _, proc := range procs {
		*ret += proc.name + "\n"
	}
	return err
}

func (r *VipmanRPC) Status(args []string, ret *string) (err error) {
	rs := startedInfo.String()
	total := 0
	for _, proc := range procs {
		total += len(proc.list)
	}
	rs += fmt.Sprintf(", started processes: %d", total)
	for _, proc := range procs {
		for _, p := range proc.list {
			pid := -1
			if p.cmd != nil {
				pid = p.cmd.Process.Pid
			}
			rs += fmt.Sprintf("\n %s [%s] %s, pid %d", proc.name, p.ip.Ip, p.cmd, pid)
		}
	}
	*ret = rs
	return err
}

func RPCClientCall(cmd, ip string, port int, args *[]string) error {
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", defaultClnAddr(ip), port))
	if err != nil {
		return err
	}
	defer client.Close()
	var ret string
	//Log("%s> %v", cmd, args)
	client.Call("VipmanRPC."+cmd, args, &ret)
	if FlagVerbose {
		fmt.Printf("RPC[%d].%s \n", port, cmd)
	}
	fmt.Println(ret)
	return nil
}

func defaultClnAddr(ip string) string {
	if len(ip) == 0 {
		return "127.0.0.1"
	} else {
		return ip
	}
}

func defaultBindAddr() string {
	return "0.0.0.0"
}

// start rpc server.
func startRpcServer(flagArgs *StartInfo, ctx context.Context, rpcChan chan<- *rpcMessage) error {
	rpc.Register(&VipmanRPC{})
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", defaultBindAddr(), flagArgs.FlagPort))
	Proxy(flagArgs)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var acceptingConns = true
	for acceptingConns {
		conns := make(chan net.Conn, 1)
		go func() {
			conn, err := server.Accept()
			if err != nil {
				return
			}
			conns <- conn
		}()
		select {
		case <-ctx.Done():
			acceptingConns = false
			break
		case client := <-conns: // server is not canceled.
			wg.Add(1)
			go func() {
				defer wg.Done()
				rpc.ServeConn(client)
			}()
		}
	}
	done := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("RPC server did not shut down in 10 seconds, quitting")
	}
}
