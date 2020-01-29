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

type proxyCmd struct {
	flagChild,
	flagProxyCmd,
	flagProxy,
	flagEth,
	flagIp string
}

func (c proxyCmd) String() string {
	return fmt.Sprintf("Child: %s, ProxyCmd: %s, FlagProxy: %s, FlagEth: %s, FlagIp: %s", c.flagChild, c.flagProxyCmd, c.flagProxy, c.flagEth, c.flagIp)
}

//If no --cmd flag, then list of proxy configurations
//If no --child flags, then will list all child proxy configurations.
//If --proxy set, then will override existing configuration` + CLIENT,

// call by vman.RPCClientCall("Proxy", FlagPort,  [2]string{flagChild, flagProxyCmd, FlagProxy, FlagEth, FlagIp})
func (r *VipmanRPC) Proxy(args []string, ret *string) (err error) {
	pc := &proxyCmd{args[0], args[1], args[2], args[3], args[4]}
	Log(pc.String())
	if len(pc.flagProxyCmd) != 0 {
		if pc.flagProxyCmd[0] == 'p' || pc.flagProxyCmd[0] == 'P' { // pause
			pauseResume(true, pc.flagChild)
		} else { // consider resume
			pauseResume(false, pc.flagChild)
		}
	} else if len(pc.flagProxy) != 0 {
		nics, err := LocalAddresses(pc.flagEth)
		if err != nil {
			return err
		}
		if len(nics) == 0 {
			return errors.New("start LocalAddresses() is empty!")
		}
		var ips []string
		for _, e := range nics {
			for _, i := range e {
				ips = append(ips, i.Ip)
			}
		}
		Proxy(&StartInfo{"", DEF_RPC_PORT, pc.flagEth, pc.flagIp, "", "", pc.flagProxy, nics, compMax(ips)})
	} else {
		list(pc.flagChild)
	}
	return nil
}

type proxyDir struct {
	Ip, PortDest, PortSrc string
	proxy                 *httputil.ReverseProxy
	pause                 bool
}

func (d proxyDir) String() interface{} {
	ret := fmt.Sprintf("%s -> %s:%s", d.PortSrc, d.Ip, d.PortDest)
	if d.pause {
		ret += " -paused"
	}
	return ret
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
				matched := 0
				paused := 0
				for _, p := range ps {
					if portDest == p.PortDest && portSrc == p.PortSrc {
						matched++
						if p.pause {
							paused++
						} else {
							proxies = append(proxies, p.proxy)
						}
					}
				}
				if len(proxies) > 0 {
					proxies[rnd.Intn(len(proxies))].ServeHTTP(w, r)
				} else if len(ps) > 0 {
					LogVerbose("No match proxy from list [%d] for map: %s -> %s ", len(ps), portSrc, portDest)
					if matched > 0 { //
						LogVerbose("Total matched: %d, but all paused", matched)
					}
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
		LogVerbose("Registering on parent %s:%d for proxying: %s %s \n", flagArgs.FlagParent, flagArgs.FlagPort, ips, flagArgs.FlagProxy)
		for _, err := RPCClientCallNoPrint("Register", flagArgs.FlagParent, flagArgs.FlagPort, &[]string{ips, flagArgs.FlagProxy}); err != nil; {
			Log("Repeating RPC call to Register Proxy after %d sec\n", SleepBeforeRepeatSec)
			time.Sleep(SleepBeforeRepeatSec * time.Second)
		}
		LogVerbose("Registered on: %s\n", flagArgs.FlagParent)
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
				}, false}
				cl = append(cl, pd)
				childs[key] = pd
				LogVerbose("Created Proxy Direct: %s -> %s:%s ", pp[0], e, portDest)
			}
		}
		proxyDirs[pp[0]] = cl
		LogVerbose("Registered reverse proxy [%d] for port [%s]", len(cl), pp[0])
	}
	*ret = fmt.Sprintf("ok [%d]", cnt)
	return nil
}

func pauseResume(pause bool, child string) {
	for k, c := range childs {
		if len(child) == 0 || (len(k) >= len(child) && k[0:len(child)-1] == child) {
			c.pause = pause
		}
	}
}

func list(child string) {
	for k, c := range childs {
		if len(child) == 0 || (len(k) >= len(child) && k[0:len(child)-1] == child) {
			fmt.Println(c.String())
		}
	}
}
