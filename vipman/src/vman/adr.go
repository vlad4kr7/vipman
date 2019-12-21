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
		fmt.Printf("localAddresses matching to [%s]\n", FlagEth)
	}
	win := runtime.GOOS == "windows"
	resp := make(map[string][]*UIP)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil || len(addrs) == 0 { // fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		if i.Name == "lo" || (len(i.Name) > 2 && i.Name[0:2] == "Lo") ||
			(len(match) > 0 && !Match(match, i.Name)) {
			continue
		}
		if win {
			callWip(i.Name)
		} else {
			resp[i.Name] = callUip(i.Name)
		}
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

func callWip(eth string) []*UIP {
	fmt.Printf("No Mindows OS to check for [%s]\n", eth)
	return nil
}

func callUip(eth string) []*UIP {
	cmd := exec.Command("ip", "a", "show", "dev", eth)
	resp := []*UIP{}
	var out bytes.Buffer
	cmd.Stdout = &out
	errcc := cmd.Run()
	if errcc != nil {
		fmt.Println(errcc)
	} else {
		for {
			str, e := out.ReadString('\n')
			if e != nil {
				break
			}
			if strings.Index(str, eth) > 0 {
				str = strings.TrimSpace(str) // , " \t\n")
				ss := strings.Split(str, " ")
				if ss[0] == "inet" {
					resp = append(resp, initUIP(eth, ss))
					//					fmt.Printf("%s\n", resp[len(resp)-1].String())
				}
			}
		}
	}
	return resp
}

// unix IP
type UIP struct {
	Id, Name, Inet, Brd, Scope, Ip, Mask string
	Secondary, Dynamic, IPv6             bool
}

func (u *UIP) String() string {
	return fmt.Sprintf("name: %s, id: %s, ip: %s / %s, brd: %s, \n\t scope: %s, v6: %t, secondary: %t, dynamic: %t",
		u.Name, u.Id, u.Ip, u.Mask, u.Scope, u.Brd, u.IPv6, u.Secondary, u.Dynamic)
}

func initUIP(eth string, ss []string) *UIP {
	var u UIP
	u.IPv6 = ss[0] == "inet6"
	if u.IPv6 {
		if ss[2] == "scope" {
			u.Scope = ss[2]
		}
		return &u
	}
	ee := ss[len(ss)-1]
	if len(ee) > len(eth) {
		ee = ee[len(eth)+1:]
	} else {
		ee = "0"
	}
	u.Id = ee
	u.Name = ss[len(ss)-1]
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
