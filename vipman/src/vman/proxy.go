package vman

import (
	"errors"
	"fmt"
	_ "golang.org/x/net/http/httpproxy"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// call by vman.RPCClientCall("Proxy", FlagPort,  [2]string{flagChild, flagProxyCmd, FlagProxy, FlagEth, FlagIp})
func (r *VipmanRPC) Proxy(args []string, ret *string) (err error) {
	flagChild := args[0]
	flagProxyCmd := args[1]
	flagProxy := args[2]
	flagEth := args[3]
	flagIp := args[4]
	Log("flagChild: %s, flagProxyCmd: %s, FlagProxy: %s, FlagEth: %s, FlagIp: %s", flagChild, flagProxyCmd, flagProxy, flagEth, flagIp)
	return errors.New("Not implemented")
}

type proxyDir struct {
	Ip, PortDest, PortSrc string
	proxy                 *httputil.ReverseProxy
}

func (d proxyDir) String() interface{} {
	return fmt.Sprintf("%s -> %s:%s", d.PortSrc, d.Ip, d.PortDest)
}

var rnd = rand.New(rand.NewSource(time.Now().Unix())) // initialize local pseudorandom generator

var proxyDirs = make(map[string][]*proxyDir)
var childs = make(map[string]*proxyDir)

// run on parent to configure proxy ports
func Proxy(flagArgs *StartInfo) {
	//flagArgs.FlagProxy
	if len(flagArgs.FlagProxy) > 0 {
		pCnt := 0
		for _, portsPairStr := range strings.Split(flagArgs.FlagProxy, ",") {
			pp := strings.Split(portsPairStr, ":")
			pCnt++
			portDest := pp[0]
			if len(pp) == 2 {
				portDest = pp[1]
			}
			ip := flagArgs.FlagIp
			if len(ip) == 0 {
				for _, e := range flagArgs.Interfaces {
					ip = e[0].Ip
					break
				}
			}
			ip += ":" + pp[0]

			// TODO add Jaeger logging
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				portSrc := r.Host[strings.Index(r.Host, ":")+1:]
				ps := proxyDirs[portSrc]
				var proxies []*httputil.ReverseProxy
				for _, p := range ps {
					if portDest == p.PortDest && portSrc == p.PortSrc {
						proxies = append(proxies, p.proxy)
					}
				}
				if len(proxies) > 0 {
					proxies[rnd.Intn(len(proxies))].ServeHTTP(w, r)
				} else if len(ps) > 0 {
					LogVerbose("No match proxy from list [%d] for map: %s -> %s", len(ps), portSrc, portDest)
				}
			})
			go servProxy(ip)
		}
		go Parent(flagArgs)
		LogVerbose("Complete proxy reg ports [%d]", pCnt)
	}
}

func servProxy(ip string) {
	LogVerbose("proxy listing [%s]", ip)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		LogVerbose("Fail proxy listing on port %s: %v\n", ip, err)
	}
}

// run on child to connect to parent RPC to register to configure and handle proxy
func Parent(flagArgs *StartInfo) {
	time.Sleep(time.Second)
	if len(flagArgs.FlagProxy) > 0 && len(procs) > 0 && len(flagArgs.FlagParent) > 0 {
		ips := ""
		for _, p := range procs {
			for _, i := range p.list {
				if len(ips) > 0 {
					ips += ","
				}
				ips += i.ip.Ip
			}
		}
		RPCClientCall("Register", flagArgs.FlagParent, flagArgs.FlagPort, &[]string{ips, flagArgs.FlagProxy})
	}
}

// called on parent to register a child
func (r *VipmanRPC) Register(args []string, ret *string) (err error) {
	Log("Registering %v", args)
	ports := args[1]
	cnt := 0
	for _, portsPairStr := range strings.Split(ports, ",") { // ports pair(s)
		pp := strings.Split(portsPairStr, ":")
		cl := proxyDirs[pp[0]]
		portDest := pp[0]
		if len(pp) == 2 {
			portDest = pp[1]
		}
		for _, e := range strings.Split(args[0], ",") { // ips
			dest := e + ":" + portDest
			key := pp[0] + ":" + dest
			pd, ok := childs[key]
			if ok {
				LogVerbose("Registering, but already present %s", pd.String())
			} else {
				pd = &proxyDir{e, portDest, pp[0], &httputil.ReverseProxy{
					Director: func(req *http.Request) {
						req.Header.Add("X-Forwarded-Host", req.Host)
						req.Header.Add("X-Origin-Host", dest)
						req.URL.Scheme = "http"
						req.URL.Host = dest
						LogVerbose("[%s] %s: %s%s -> %s\n", req.Proto, req.Method, req.Host, req.RequestURI, dest)
					},
				}}
				cl = append(cl, pd)
				LogVerbose("Created Proxy Direct: %s -> %s:%s ", pp[0], e, portDest)
			}
		}
		proxyDirs[pp[0]] = cl
		LogVerbose("Registered reverse proxy [%d] for port [%s]", len(cl), pp[0])
	}
	*ret = fmt.Sprintf("ok [%d]", cnt)
	return nil
}

func pause() {
	Panic("")

}

func resume() {
	Panic("")

}

func list() {
	Panic("")

}
