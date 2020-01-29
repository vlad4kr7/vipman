package vman

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func Prepare(flagEth, flagSet string, flagClean bool) {
	if len(flagSet) == 0 && !flagClean {
		prepareShow(flagEth)
	} else if runtime.GOOS != "linux" {
		Panic("prepare command use 'ip link add' and 'ip show' which is available only on %s\n", "linux")
	} else if len(flagSet) != 0 && flagClean {
		Panic("--set and --clean NOT ALLOWED! Check help: '%s prepare -h' for details\n", "vipman")
	} else if len(flagSet) != 0 {
		prepareAdd(flagEth, flagSet)
		prepareShow(flagEth)
	} else if flagClean {
		prepareDelete(flagEth)
		prepareShow(flagEth)
	}
}

func prepareShow(flagEth string) {
	lst, err := LocalAddresses(flagEth)
	if err != nil {
		Log("LocalAddresses(): ", err)
	} else {
		for k, e := range lst {
			if len(e) > 0 {
				Log("Ethernet: %s [%d]", k, len(e))
				for _, i := range e {
					Log(i.String())
				}
			}
		}
	}
}

func prepareAdd(flagEth, flagSet string) {
	lst, err := LocalAddresses(flagEth)
	if err != nil {
		Panic("PrepareAdd: LocalAddresses: Err: %v\n", err)
	} else {
		max, err := strconv.Atoi(flagSet)
		if err != nil {
			Panic("PrepareAdd: FlagAdd: %v\n", err)
			return
		}
		for n, e := range lst {
			if len(e) == 0 {
				Panic("No primary adapter for %s\n", n)
			}
			if e[0].IdNotParsed {
				Panic("'ip link add ...%s' NOR 'dhclient %s' wont create an IP alias on WSL!\n", n, n)
				continue
			}
			for i := len(e); i <= max; i++ {
				etha := n + "_" + strconv.Itoa(i)
				cmd := exec.Command("ip", "link", "add", etha, "link", n, "type", "macvlan", "mode", "bridge")
				var out bytes.Buffer
				cmd.Stdout = &out
				var outerr bytes.Buffer
				cmd.Stderr = &outerr
				if FlagVerbose {
					out.WriteString("IP alias added by command: ip link add " + etha)
				}

				cmdErr := false
				err1 := cmd.Run()
				if err1 != nil {
					//					LogError("PrepareAdd: %s\nSuggestion: consider to run as 'sudo vipman..'", etha)
					cmdErr = true
				} else {
					if FlagVerbose {
						out.WriteString("IP alias configured by command: dhclient " + etha)
					}

					cmd2 := exec.Command("dhclient", etha)
					cmd2.Stdout = &out
					cmd2.Stderr = &outerr
					err2 := cmd2.Run()
					if err2 != nil {
						LogError("PrepareAdd dhclient: %v", err2.Error())
						cmdErr = true
					} else {
						if FlagVerbose {
							out.WriteString("IP alias started by command: ifconfig " + etha + " up")
						}

						cmd3 := exec.Command("ifconfig", etha, "up")
						cmd3.Stdout = &out
						cmd3.Stderr = &outerr
						err3 := cmd3.Run()
						if err3 != nil {
							//							LogError("PrepareAdd ifconfig up: %v", err2.Error())
							cmdErr = true
						}
					}
				}
				if FlagVerbose && out.Len() > 0 {
					Log("out> " + out.String())
				}
				if outerr.Len() > 0 {
					Log("err> " + outerr.String())
				}
				if cmdErr {
					os.Exit(1)
				}
			}
		}
	}
}

func prepareDelete(flagEth string) {
	lst, err := LocalAddresses(flagEth)
	if err != nil {
		Panic("prepareDelete: LocalAddresses: Err: %v\n", err)
	} else {
		if err != nil {
			Panic("prepareDelete:  %v\n", err)
			return
		}
		for n, e := range lst {
			if len(e) == 0 {
				Panic("No primary adapter for %s\n", n)
			}
			if e[0].IdNotParsed {
				Panic("'ip a add ...%s' NOR 'dhclient %s' wont create an IP alias on WSL!\n", n, n)
				continue
			}
			for i := len(e) - 1; i > 0; i-- {
				etha := e[i].Name
				if strings.IndexByte(etha, '_') < 0 {
					continue
				}
				cmd := exec.Command("ip", "link", "delete", etha)
				var out bytes.Buffer
				cmd.Stdout = &out
				var outerr bytes.Buffer
				cmd.Stderr = &outerr
				err1 := cmd.Run()
				if err1 != nil {
					Panic("prepareDelete: %s\nSuggestion: consider to run as 'sudo vipman..'", etha)
				}
				if FlagVerbose {
					out.WriteString("IP alias cleaned by command: ip link delete " + etha)
				}
				if FlagVerbose && out.Len() > 0 {
					Log("out> " + out.String())
				}
				if outerr.Len() > 0 {
					Log("err> " + outerr.String())
				}
			}
		}
	}
}
