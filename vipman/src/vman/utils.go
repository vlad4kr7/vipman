package vman

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strings"
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
	if len(msg) == 0 {
		msg = "NOT IMPLEMENTED YET\n"
	}
	log.SetOutput(os.Stderr)
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

// not used
func examiner(t reflect.Type, depth int) {
	fmt.Println(strings.Repeat("\t", depth), "Type is", t.Name(), "and kind is", t.Kind())
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		fmt.Println(strings.Repeat("\t", depth+1), "Contained type:")
		examiner(t.Elem(), depth+1)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Println(strings.Repeat("\t", depth+1), "Field", i+1, "name is", f.Name, "type is", f.Type.Name(), "and kind is", f.Type.Kind())
			if f.Tag != "" {
				fmt.Println(strings.Repeat("\t", depth+2), "Tag is", f.Tag)
				fmt.Println(strings.Repeat("\t", depth+2), "tag1 is", f.Tag.Get("tag1"), "tag2 is", f.Tag.Get("tag2"))
			}
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compMax2(s1, s2 string) int {
	max := 0
	l1 := len(s1) - 1
	l2 := len(s2) - 1
	ml := minInt(l1, l2)
	for i := 0; i < ml-1; i++ {
		if s1[i] != s2[i] {
			max = i
			break
		}
	}
	if max == 0 && (s1[ml] != s2[ml] || l1 == l2) {
		max = ml
	}
	return max
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

func compMaxIp(ips []string) int {
	if len(ips) < 2 {
		return 0
	}
	m := make(map[int]map[string]int)
	max := 0
	maxi := 0
	for _, i := range ips {
		if maxi < len(i) {
			maxi = len(i)
		}
		for ax, a := range strings.Split(i, ".") {
			lst, ok := m[ax]
			if !ok {
				lst = make(map[string]int)
			}
			lst[a] = lst[a] + 1
			m[ax] = lst
			//fmt.Printf("comp %v", lst)
		}
	}
	for i, lst := range m {
		if len(lst) > 1 {
			break
		}
		for ip, _ := range lst {
			max += len(ip) + 1
			//if i < len(m)-1 {
			//	max ++
			//}
			s := fmt.Sprintf("%d %s,  max= %d %d %d", i, ip, max, len(ip), len(m))
			fmt.Println(s)
			break
		}
	}
	if max >= maxi {
		max = 0
	}
	return max
}
