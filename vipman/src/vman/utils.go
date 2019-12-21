package vman

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var FlagEth, FlagProcfile, FlagBaseDir, FlagPort, FlagParent, FlagAdd, FlagIp string
var FlagClean, FlagVerbose bool

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

func Panic(msg string) {
	if len(msg) == 0 {
		msg = "NOT IMPLEMENTED YET\n"
	}
	fmt.Fprintf(os.Stderr, msg)
	os.Exit(1)
}

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
