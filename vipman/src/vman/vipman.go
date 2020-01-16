package vman

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	_ "golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	_ "runtime"
	"strings"
	"sync"
	"syscall"
)

// -- process information structure.
type procInfo struct {
	name    string
	cmdline string
	cmd     *exec.Cmd

	ip *UIP
	//	port       uint
	//	setPort    bool

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
func readProcfile() error {
	content, err := ioutil.ReadFile(FlagProcfile)
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

func findProc(name string) *procInfo {
	mu.Lock()
	defer mu.Unlock()

	for _, proc := range procs {
		if proc.name == name {
			return proc
		}
	}
	return nil
}

//var colors []string
/*
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
*/

func Start() {
	err := readProcfile()
	if err != nil {
		Panic("readProcfile", err)
	}
	c := notifyCh()
	start(context.Background(), c)
}

func start(ctx context.Context, sig <-chan os.Signal) error {
	/*
		err := readProcfile(cfg)
		if err != nil {
			return err
		}
	*/
	ctx, cancel := context.WithCancel(ctx)
	// Cancel the RPC server when procs have returned/errored, cancel the
	// context anyway in case of early return.
	defer cancel()

	/*
		if len(cfg.Args) > 1 {
			tmp := make([]*procInfo, 0, len(cfg.Args[1:]))
			maxProcNameLength = 0
			for _, v := range cfg.Args[1:] {
				proc := findProc(v)
				if proc == nil {
					return errors.New("unknown proc: " + v)
				}
				tmp = append(tmp, proc)
				if len(v) > maxProcNameLength {
					maxProcNameLength = len(v)
				}
			}

			mu.Lock()
			procs = tmp
			mu.Unlock()
		}
	*/

	godotenv.Load()
	rpcChan := make(chan *rpcMessage, 10)
	go startRpcServer(ctx, rpcChan)

	procsErr := startAllProcs(sig, rpcChan, true)
	return procsErr
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

const sigint = syscall.SIGINT
const sigterm = syscall.SIGTERM
const sighup = syscall.SIGHUP

func notifyCh() <-chan os.Signal {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, sigterm, sigint, sighup)
	return sc
}

var cmdStart = []string{"/bin/sh", "-c"}

//var procAttrs = &unix.SysProcAttr{Setpgid: true}

func terminateProc(proc *procInfo, signal os.Signal) error {
	return proc.cmd.Process.Signal(signal)
	/*
		p := proc.cmd.Process
		if p == nil {
			return nil
		}

		pgid, err := unix.Getpgid(p.Pid)
		if err != nil {
			return err
		}

		// use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
		pid := p.Pid
		if pgid == p.Pid {
			pid = -1 * pid
		}

		target, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		return target.Signal(signal)
	*/
}

// killProc kills the proc with pid pid, as well as its children.
//func killProc(process *os.Process) error {
//	return unix.Kill(-1*process.Pid, unix.SIGKILL)
//}
