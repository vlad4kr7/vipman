package vman

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
)

func Prepare() {
	//    if len(FlagEth) ==0 &&
	if len(FlagAdd) == 0 && !FlagClean {
		PrepareShow()
	} else if len(FlagAdd) != 0 && FlagClean {
		Panic("NOT ALLOWED")
	} else if len(FlagAdd) != 0 {
		PrepareAdd()
		PrepareShow()
	} else if FlagClean {
		Panic("")
	}
}

func PrepareShow() {
	lst, err := LocalAddresses(FlagEth)
	if err != nil {
		fmt.Printf("Err: %v", err)
	} else {
		for k, e := range lst {
			if len(e) > 0 {
				fmt.Printf("Ethernet: %s [%d]\n", k, len(e))
				for _, i := range e {
					fmt.Printf("\t%s\n", i.String())
				}
			}
		}
	}
}

func PrepareAdd() {
	lst, err := LocalAddresses(FlagEth)
	if err != nil {
		fmt.Printf("PrepareAdd: LocalAddresses: Err: %v\n", err)
	} else {
		max, err := strconv.Atoi(FlagAdd)
		if err != nil {
			fmt.Printf("PrepareAdd: FlagAdd: %v\n", err)
			return
		}
		for n, e := range lst {
			for i := len(e); i <= max; i++ {
				etha := n + ":" + strconv.Itoa(i)

				cmd := exec.Command("dhclient", "-4", etha)
				var out bytes.Buffer
				cmd.Stdout = &out
				var outerr bytes.Buffer
				cmd.Stderr = &outerr

				errr := cmd.Run()
				if errr != nil {
					fmt.Printf("PrepareAdd: dhclient %s\n%s\nSuggestion: consider to run as 'sudo vipman..'", etha, out.String(), outerr.String())
				}
			}
		}
	}
}
