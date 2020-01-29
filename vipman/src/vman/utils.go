package vman

import (
	"log"
	"math"
	"os"
)

// default RPC server port
const DEF_RPC_PORT int = 17654

var FlagVerbose, FlogTime bool

func Match(a, b string) bool {
	lena := len(a) - 1
	if len(b) > 0 && lena >= 0 && len(b) >= lena && a[lena] == '*' {
		//		Log("%t %t \n", a[0:lena], b[0:lena])
		return a[0:lena] == b[0:lena]
	} else {
		//		Log("%s = %s \n", a, b)
		return a == b
	}
}

func Panic(msg string, v ...interface{}) {
	log.Fatalf(msg, v...)
}

func Log(msg string, v ...interface{}) {
	if msg[len(msg)-1] != '\n' {
		msg = msg + "\n"
	}
	log.Printf(msg, v...)
}

func LogError(msg string, v ...interface{}) {
	if msg[len(msg)-1] != '\n' {
		msg = msg + "\n"
	}
	log.SetOutput(os.Stderr)
	log.Printf(msg, v...)
	log.SetOutput(os.Stdout)
}

func LogVerbose(msg string, v ...interface{}) {
	if FlagVerbose {
		Log(msg, v...)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compMax20(s1, s2 string) int {
	max := 0
	for i := 0; i < minInt(len(s1), len(s2)); i++ {
		if s1[i] != s2[i] {
			max = i
			break
		}
	}
	return max
}

func compMax(ips []string) int {
	if len(ips) < 2 {
		return 0
	}
	max := math.MaxInt16
	for i := 1; i < len(ips); i++ {
		m := compMax20(ips[i-1], ips[i])
		if m > 0 && m < max {
			max = m
		}
	}
	if max == math.MaxInt16 {
		max = 0
	}
	return max
}
