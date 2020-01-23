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
	LogVerbose("localAddresses matching to [%s] on OS: %s", match, runtime.GOOS)
	resp := make(map[string][]*UIP)
	for _, i := range ifaces {
		//i.HardwareAddr.String()
		addrs, err := i.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		if i.Name == "lo" || (len(i.Name) > 2 && i.Name[0:2] == "Lo") ||
			(len(match) > 0 && !Match(match, i.Name)) {
			continue
		}

		if runtime.GOOS != "linux" {
			Log("non linux use - IP addresses information in incomplete")
			callGoIp(i, resp)
		} else {
			callUip(i.Name, resp)
		}
	}
	return resp, nil
}

func callGoIp(nic net.Interface, mapa map[string][]*UIP) {
	adders, err := nic.Addrs()
	if err != nil {
		Panic(err.Error())
	}
	// handle err
	var resp []*UIP
	for i, addr := range adders {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			{
				ip = v.IP
				if v.IP.To4() == nil || v.IP.IsLoopback() {
					continue
				}
			}
		case *net.IPAddr:
			ip = v.IP
		}
		// process IP address
		resp = append(resp, initGoIP(i, nic.Name, ip))
	}
	mapa[nic.Name] = resp
	/*
		for _, a := range adders {
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

func initGoIP(id int, name string, eth net.IP) *UIP {
	var u UIP
	u.Name = name
	u.Id = fmt.Sprint(id)
	u.Ip = eth.To4().String()
	u.Mask = eth.DefaultMask().String()
	u.NotLinux = true
	return &u
}

func callUip(eth string, mapa map[string][]*UIP) { //(string, []*UIP) {
	cmd := exec.Command("ip", "a", "show", "dev", eth)
	LogVerbose("exec %s", cmd.String())
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
	Id, Name, Inet, Brd, Scope, Ip, Mask            string
	Secondary, Dynamic, IPv6, IdNotParsed, NotLinux bool
}

func (u *UIP) String() string {
	id := ""
	if u.Id != "0" {
		id = fmt.Sprintf(" # id: %s", u.Id)
		if u.IdNotParsed {
			id += " (IdNotParsed)"
		}
	}
	resp := fmt.Sprintf("name: %s%s, ip: %s ", u.Name, id, u.Ip)
	if !u.NotLinux {
		resp += fmt.Sprintf("/ %s, brd: %s,\n\t scope: %s, v6: %t, secondary: %t, dynamic: %t",
			u.Mask, u.Brd, u.Scope, u.IPv6, u.Secondary, u.Dynamic)
	}
	return resp
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
	u.NotLinux = false
	return &u
}
