package vman

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/spf13/cobra"
	_ "golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	_ "os/signal"
	"regexp"
	_ "runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

type proc struct {
	ip                  *UIP
	colorIndex          int
	cmd                 *exec.Cmd
	pi                  *procInfo
	stoppedBySupervisor bool
	cmdLine             string // as on start
	errCh               chan<- error
	logger              *clogger
	mu                  sync.Mutex
}

func (p proc) appName() string {
	return strings.SplitN(p.cmdLine, " ", 1)[0]
}

// -- process information structure.
type procInfo struct {
	name                string
	cmdline             string // as in Proc file
	list                []*proc
	stoppedBySupervisor bool
	mainColorIndex      int
	ips                 string // comma separated list of used IPs
}

func (i procInfo) logName(f *StartInfo, ip *UIP) string {
	return fmt.Sprintf("%s [%s]", i.name, ip.Ip[f.NameDiff:])
}

var mu sync.Mutex

// process informations named with proc.
var procs []*procInfo

var maxProcNameLength = 0

var re = regexp.MustCompile(`\$([a-zA-Z]+[a-zA-Z0-9_]+)`)

type StartInfo struct {
	FlagProcfile                string
	FlagPort                    int
	FlagEth, FlagIp, FlagParent string
	FlagBaseDir, FlagProxy      string
	FlagWorkspaceDir            string
	Interfaces                  map[string][]*UIP
	NameDiff                    int
	Ips                         string
}

func (si *StartInfo) String() string {
	return fmt.Sprintf("Procfile: %s, RPC port: %d, eth: %s, ip: %s, parent: %s, baseDir: %s, proxy: %s", si.FlagProcfile, si.FlagPort, si.FlagEth, si.FlagIp, si.FlagParent, si.FlagBaseDir, si.FlagProxy)
}

var startedInfo *StartInfo

// read Procfile and parse it.
func readProcfile(flagProcfile string) {
	content, err := ioutil.ReadFile(flagProcfile)
	if err != nil {
		Panic("readProcfile: %s\n", err.Error())
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

		proc := &procInfo{name: k, cmdline: v, mainColorIndex: index, list: []*proc{}}

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
		Panic("%s: no valid entry\n", "readProcfile")
	}
}

// all procs OR active Only process
func totalProcs(all bool) int {
	i := 0
	mu.Lock()
	defer mu.Unlock()
	for _, proc := range procs {
		for _, p := range proc.list {
			if all || p.cmd != nil {
				i++
			}
		}
	}
	return i
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

func Start(flagArgs *StartInfo) {
	startedInfo = flagArgs
	readProcfile(flagArgs.FlagProcfile)
	sigChan := notifyCh()
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel the RPC server when procs have returned/errored, cancel the
	// context anyway in case of early return.
	defer cancel()
	godotenv.Load()
	rpcChan := make(chan *rpcMessage, 10)
	if len(flagArgs.FlagIp) > 0 {
		flagArgs.Ips = flagArgs.FlagIp
	} else {
		nics, err := LocalAddresses(flagArgs.FlagEth)
		if err != nil {
			Panic("start LocalAddresses(): %s \n", err.Error())
		}
		if len(nics) == 0 {
			Panic("%s LocalAddresses() is empty!\n", "start")
		}
		flagArgs.Interfaces = nics
		ipss := ""
		var ips []string
		for _, e := range nics {
			for _, i := range e {
				ips = append(ips, i.Ip)
				ipss += i.Ip + ","
			}
		}
		flagArgs.NameDiff = compMax(ips)
		flagArgs.Ips = ipss[0 : len(ipss)-2]
	}
	go startRpcServer(flagArgs, ctx, rpcChan)
	startAllProcs(flagArgs, sigChan, rpcChan)
}

func (r *VipmanRPC) Stop(args []string, ret *string) (err error) {
	flagIp := args[0]
	flagProcName := args[1]
	*ret = ""
	for _, proc := range procs {
		if proc.name == flagProcName {
			for _, p := range proc.list {
				if flagIp == p.ip.Ip && p.cmd != nil {
					*ret += "stopping: " + p.appName()
					terminateProc(p, os.Interrupt)
				}
			}
		}
	}
	return nil
}

func (r *VipmanRPC) StopAll(args []string, ret *string) (err error) {
	err = stopProcs(os.Interrupt)
	if err != nil {
		*ret = err.Error()
	} else {
		*ret = "stopped"
	}
	return err
}

func (r *VipmanRPC) Restart(args []string, ret *string) (err error) {
	flagIp := args[0]
	flagProcName := args[1]
	*ret = ""
	for _, proc := range procs {
		if proc.name == flagProcName {
			for _, p := range proc.list {
				if flagIp == p.ip.Ip && p.cmd != nil {
					*ret += "restarting: " + p.appName()
					go spawnProc(p, true)
				}
			}
		}
	}
	return errors.New(*ret)
}

const sigint = syscall.SIGINT
const sigterm = syscall.SIGTERM
const sighup = syscall.SIGHUP

// Register system interrupt chanel
func notifyCh() <-chan os.Signal {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, sigterm, sigint, sighup)
	return sc
}

// spawnProc starts the specified proc, and returns any error from running it.
func spawnProc(p *proc, kill bool) {
	if kill && p.cmd != nil {
		err := terminateProc(p, os.Interrupt)
		if err != nil {
			Log("Wont stop - wont start %s", err.Error())
		}
	}
	p.mu.Lock()
	cs := strings.Split(p.cmdLine, " ")
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = p.logger
	cmd.Stderr = p.logger
	//	cmd.SysProcAttr = procAttrs
	cmd.Env = append(os.Environ(), fmt.Sprintf("IP=%s", p.ip.Ip))
	cmd.Env = append(cmd.Env, fmt.Sprintf("IPS=%s", p.pi.ips))
	cmd.Env = append(cmd.Env, fmt.Sprintf("WORKSPACE=%s", p.pi.ips))
	fmt.Fprintf(p.logger, "Starting %s %s on %s\n", p.pi.name, p.cmdLine, p.ip.Ip)
	if err := cmd.Start(); err != nil {
		select {
		case p.errCh <- err:
		default:
		}
		fmt.Fprintf(p.logger, "Failed to start %s: %s\n", p.pi.name, err)
		p.mu.Unlock()
		return
	}
	p.cmd = cmd
	p.mu.Unlock()
	err := cmd.Wait()
	p.cmd = nil
	if err != nil {
		fmt.Fprintf(p.logger, "Terminating %s\n", err)
	} else {
		err = errors.New("stub")
	}
	fmt.Fprint(p.logger, "Terminated\n")
	p.errCh <- err
}

func stopProcByName(name string) error {
	Log("stopProcByName: %s", name)
	return stopProc(findProc(name), os.Interrupt)
}

// StopRC the specified proc, issuing os.Kill if it does not terminate within 10
// seconds. If signal is nil, os.Interrupt is used.
func stopProc(proc *procInfo, signal os.Signal) error {
	if signal == nil {
		signal = os.Interrupt
	}
	if proc == nil {
		return errors.New("unknown proc to stop")
	}
	if len(proc.list) == 0 {
		return nil
	}
	proc.stoppedBySupervisor = true

	var em []error
	for _, p := range proc.list {
		p.stoppedBySupervisor = true

		err := terminateProc(p, signal)
		if err != nil {
			em = append(em, err)
		}
	}
	if len(em) > 0 {
		emMsg := ""
		for i, e := range em {
			if i > 0 {
				emMsg += ", "
			}
			emMsg += e.Error()
		}
		return errors.New(emMsg)
	}

	return nil
}

// start specified proc. if proc is started already, return nil.
func startProc(flagArgs *StartInfo, id int, pi *procInfo, ip *UIP, errCh chan<- error) {
	p := &proc{ip: ip, colorIndex: pi.mainColorIndex, pi: pi, errCh: errCh}
	pi.list = append(pi.list, p)
	p.logger = createLogger(p.pi.logName(flagArgs, ip), p.colorIndex, id)
	cmdline := strings.ReplaceAll(p.pi.cmdline, "$IP", p.ip.Ip) // set process IP
	cmdline = strings.ReplaceAll(cmdline, "$IPS", flagArgs.Ips) // set list of used IP by all processes
	p.pi.ips = flagArgs.Ips
	if !(cmdline[0] == '.' || cmdline[0] == '/') {
		cmdline = flagArgs.FlagBaseDir + "/" + cmdline
	}
	p.cmdLine = cmdline
	go spawnProc(p, false)
}

// stopProcs attempts to stop every running process and returns any non-nil
// error, if one exists. stopProcs will wait until all procs have had an
// opportunity to stop.
func stopProcs(sig os.Signal) error {
	var err error
	Log("Stopping all [%d] procs", len(procs))
	for _, proc := range procs {
		stopErr := stopProc(proc, sig)
		if stopErr != nil {
			err = stopErr
		}
	}
	return err
}

// start all procs.
func startAllProcs(flagArgs *StartInfo, sc <-chan os.Signal, rpcCh <-chan *rpcMessage) {
	errCh := make(chan error, 1)
	if len(flagArgs.FlagIp) > 0 {
		for _, proc := range procs {
			ip := &UIP{Ip: flagArgs.FlagIp}
			startProc(flagArgs, 0, proc, ip, errCh)
		}
	} else {
		cnt := 0
		for _, ips := range flagArgs.Interfaces {
			for _, ip := range ips {
				for _, proc := range procs {
					startProc(flagArgs, cnt, proc, ip, errCh)
					cnt++
				}
			}
		}
	}
	time.Sleep(time.Second)
	for totalProcs(false) > 0 {
		_ = <-errCh
	}
}

func terminateProc(proc *proc, signal os.Signal) error {
	proc.mu.Lock()
	defer proc.mu.Unlock()
	if proc.cmd == nil || proc.cmd.ProcessState.Exited() {
		proc.cmd = nil
		return nil
	}
	err := proc.cmd.Process.Signal(signal)
	if err != nil {
		return err
	}
	timeout := time.AfterFunc(10*time.Second, func() {
		err = terminateProc(proc, syscall.SIGKILL)
	})
	timeout.Stop()
	return err
}
