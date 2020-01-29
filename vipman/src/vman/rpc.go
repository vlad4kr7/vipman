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

func (r *VipmanRPC) List(args []string, ret *string) (err error) {
	*ret = ""
	flagPort := DEF_RPC_PORT
	for _, proc := range childs {
		*ret += "[" + proc.Ip + "]\n" + RPCClientCallNoPrint("Status", proc.Ip, flagPort, 0, &[]string{}) + "\n"
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
	Log(RPCClientCallNoPrint(cmd, "", port, 0, &[]string{}))
}

func RPCClientCall(cmd, ip string, port int, args *[]string) {
	Log(RPCClientCallNoPrint(cmd, ip, port, 0, args))
}

const SleepBeforeRepeatSec = 10

func RPCClientCallNoPrint(cmd, ip string, port, repeat int, args *[]string) string {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", defaultClnAddr(ip), port), SleepBeforeRepeatSec/2*time.Second)
	if err != nil {
		Log("ERROR dialing %s %s:%d : %v \n", cmd, ip, port, err)
		if repeat > 0 {
			Log("Repeating RPC call %s after %d sec\n", cmd, SleepBeforeRepeatSec)
			time.Sleep(SleepBeforeRepeatSec * time.Second)
			return RPCClientCallNoPrint(cmd, ip, port, repeat-1, args)
		} else {
			return err.Error()
		}
	}
	client := rpc.NewClient(conn)
	defer client.Close()
	var ret string
	errc := client.Call("VipmanRPC."+cmd, args, &ret)
	if errc != nil {
		Log("RPC Call VipmanRPC.%s (%v) ERROR: %s\n", cmd, args, errc)
		return errc.Error()
	}
	if FlagVerbose {
		fmt.Printf("RPC[%d].%s \n", port, cmd)
	}
	return ret
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
