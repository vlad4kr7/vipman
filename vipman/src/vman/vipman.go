package vman

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"regexp"
	_ "runtime"
	"strings"
	"sync"
)

// -- process information structure.
type procInfo struct {
	name       string
	cmdline    string
	cmd        *exec.Cmd
	port       uint
	setPort    bool
	colorIndex int

	// True if we called stopProc to kill the process, in which case an
	// *os.ExitError is not the fault of the subprocess
	stoppedBySupervisor bool

	mu      sync.Mutex
	cond    *sync.Cond
	waitErr error
}

var mu sync.Mutex

// process informations named with proc.
var procs []*procInfo

var maxProcNameLength = 0

var re = regexp.MustCompile(`\$([a-zA-Z]+[a-zA-Z0-9_]+)`)

// read Procfile and parse it.
func readProcfile(cfg *config) error {
	content, err := ioutil.ReadFile(cfg.Procfile)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()

	procs = []*procInfo{}
	index := 0
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}

		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		/*
			if runtime.GOOS == "windows" {
				v = re.ReplaceAllStringFunc(v, func(s string) string {
					return "%" + s[1:] + "%"
				})
			}
		*/
		proc := &procInfo{name: k, cmdline: v, colorIndex: index}
		/*
			if *setPorts == true {
				proc.setPort = true
				proc.port = cfg.BasePort
				cfg.BasePort += 100
			}
		*/
		proc.cond = sync.NewCond(&proc.mu)
		procs = append(procs, proc)
		if len(k) > maxProcNameLength {
			maxProcNameLength = len(k)
		}
		index++
		if index >= len(colors) {
			index = 0
		}
	}
	if len(procs) == 0 {
		return errors.New("no valid entry")
	}
	return nil
}

var colors []string

type config struct {
	Procfile string `yaml:"procfile"`
	// Port for RPC server
	Port     uint   `yaml:"port"`
	BaseDir  string `yaml:"basedir"`
	BasePort uint   `yaml:"baseport"`
	Args     []string
	// If true, exit the supervisor process if a subprocess exits with an error.
	ExitOnError bool `yaml:"exit_on_error"`
}

func Start() {
	Panic("")

}
func Stop() {
	Panic("")

}
func StopAll() {
	Panic("")

}
func Restart() {
	Panic("")

}

func RestartAll() {
	Panic("")

}
