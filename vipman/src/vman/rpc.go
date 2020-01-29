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

//List suppress args
func (r *VipmanRPC) List(args []string, ret *string) (err error) {
	*ret = ""
	flagPort := DEF_RPC_PORT
	for _, proc := range childs {
		str, err := RPCClientCallNoPrint("Status", proc.Ip, flagPort, &[]string{})
		if err != nil {
			str = err.Error()
		}
		*ret += "[" + proc.Ip + "]\n" + str + "\n"
	}
	fmt.Println(ret)
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

func RPCClientCallNA(cmd string, port int) {
	Log(RPCClientCallNoPrint(cmd, "", port, &[]string{}))
}

func RPCClientCall(cmd, ip string, port int, args *[]string) {
	Log(RPCClientCallNoPrint(cmd, ip, port, args))
}

const SleepBeforeRepeatSec = 10

func RPCClientCallNoPrint(cmd, ip string, port int, args *[]string) (string, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", defaultClnAddr(ip), port), SleepBeforeRepeatSec/2*time.Second)
	if err != nil {
		Log("ERROR dialing %s %s:%d : %v \n", cmd, ip, port, err)
		return "", err
	}
	client := rpc.NewClient(conn)
	defer client.Close()
	var ret string
	errc := client.Call("VipmanRPC."+cmd, args, &ret)
	if errc != nil {
		Log("RPC Call VipmanRPC.%s (%v) ERROR: %s\n", cmd, args, errc)
		return "", errc
	}
	if FlagVerbose {
		fmt.Printf("RPC[%d].%s \n", port, cmd)
	}
	return ret, nil
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
	err := rpc.Register(&VipmanRPC{})
	if err != nil {
		return err
	}
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
