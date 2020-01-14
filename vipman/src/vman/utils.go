package vman

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

var FlagEth, FlagProcfile, FlagBaseDir, FlagPort, FlagParent, FlagSet, FlagIp string
var FlagClean, FlagVerbose bool

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
