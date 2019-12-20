package vman

import (
	"fmt"
	//	"net"
	//	"os/exec"
)

func PrepareAdd(args ...string) {
	fmt.Printf("NOT DONE %v\n", args)
	//	exec.Command("")
}

func Match(a, b string) bool {
	//fmt.Printf("%t %t \n", len(a)>=len(b),a[len(a)-1]=='*')
	lena := len(a) - 1
	if len(b) > 0 && lena >= 0 && len(b) >= lena && a[lena] == '*' {
		//fmt.Printf("%s[%s] = %s[%s] \n", a, a[0:lena], b, b[0:lena])
		return a[0:lena] == b[0:lena]
	} else {
		return a == b
	}
}
