package vman

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func LocalAddresses(match string) (map[string][]*UIP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	if FlagVerbose {
		Log("localAddresses matching to [%s] on OS: %s", FlagEth, runtime.GOOS)
	}
	resp := make(map[string][]*UIP)
	for _, i := range ifaces {
		i.HardwareAddr.String()
		addrs, err := i.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		if i.Name == "lo" || (len(i.Name) > 2 && i.Name[0:2] == "Lo") ||
			(len(match) > 0 && !Match(match, i.Name)) {
			continue
		}
		callUip(i.Name, resp)
		/*
			for _, a := range addrs {
				//			examiner(reflect.TypeOf(a), 0)
				switch v := a.(type) {
				case *net.IPNet: //IPAddr:
					if v.IP.To4() == nil || v.IP.IsLoopback() {
						continue
					}
					tx, _ := v.IP.MarshalText()
					fmt.Printf("%v : %s (%s) u?%t %s %d\n", i.Name, v, v.IP.String(),
						v.IP.IsUnspecified(), string(tx), len(tx))
				}

			}
		*/
	}
	return resp, nil
}

func callUip(eth string, mapa map[string][]*UIP) { //(string, []*UIP) {
	cmd := exec.Command("ip", "a", "show", "dev", eth)
	if FlagVerbose {
		Log("exec %s", cmd.String())
	}
	//	resp := []*UIP{}
	var out bytes.Buffer
	cmd.Stdout = &out
	errcc := cmd.Run()
	if errcc != nil {
		Panic("ip show", errcc)
	} else {
		name := eth
		var resp []*UIP
		for {
			str, e := out.ReadString('\n')
			if e != nil {
				break
			}
			ss := strings.Split(strings.TrimSpace(str), " ")
			if ss[0] == "inet" {
				resp = append(resp, initUIP(len(resp), eth, ss))
				mapa[name] = resp
				//	fmt.Printf("%s\n", resp[len(resp)-1].String())
			} else if len(ss) > 1 {
				ss1 := ss[1]
				ni := strings.Index(ss[1], "@")
				if ni > 0 {
					name = ss1[ni+1 : len(ss1)-1]
					if FlagVerbose {
						Log("Parent interface %s ", name)
					}
					resp = mapa[name]
				}
			}
		}
	}
}

// unix IP
type UIP struct {
	Id, Name, Inet, Brd, Scope, Ip, Mask  string
	Secondary, Dynamic, IPv6, IdNotParsed bool
}

func (u *UIP) String() string {
	id := ""
	if u.Id != "0" {
		id = fmt.Sprintf(" # id: %s", u.Id)
		if u.IdNotParsed {
			id += " (IdNotParsed)"
		}
	}
	return fmt.Sprintf("name: %s%s, ip: %s / %s, brd: %s, \n\t scope: %s, v6: %t, secondary: %t, dynamic: %t",
		u.Name, id, u.Ip, u.Mask, u.Brd, u.Scope, u.IPv6, u.Secondary, u.Dynamic)
}

func initUIP(id int, eth string, ss []string) *UIP {
	var u UIP
	u.IPv6 = ss[0] == "inet6"
	if u.IPv6 {
		if ss[2] == "scope" {
			u.Scope = ss[2]
		}
		return &u
	}
	ee := ss[len(ss)-1]
	if ee == eth {
		ee = "0"
		u.IdNotParsed = false
	} else if len(ee) > len(eth) && strings.Index(ee, eth) >= 0 {
		ee = ee[len(eth)+1:]
		u.IdNotParsed = false
	} else {
		ee = fmt.Sprintf("%d", id)
		u.IdNotParsed = true
	}
	u.Id = ee
	if strings.Index(ee, eth) >= 0 {
		u.Name = ss[len(ss)-1]
	} else {
		u.Name = eth
	}
	u.Inet = ss[1]
	sl := strings.Index(ss[1], "/")
	if sl > 0 {
		u.Ip = u.Inet[0:sl]
		u.Mask = u.Inet[sl+1:]
	} else {
		u.Ip = u.Inet
		u.Mask = "32"
	}
	shift := 3
	if ss[2] == "brd" {
		u.Brd = ss[3]
		shift += 2
	} else {
		u.Brd = ""
	}
	u.Scope = ss[shift]
	u.Secondary = ss[shift+1] == "secondary"
	u.Dynamic = ss[shift+1] == "dynamic"

	return &u
}
