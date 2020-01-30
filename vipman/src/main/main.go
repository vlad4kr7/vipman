package main

import (
	"dbinvent.com/vipman/vman"
	"github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	"os"
)

type CommandVO struct {
	CobraCommand  *cobra.Command
	FlagFunctions []func(*cobra.Command) string
	RequiredFlags []int
}

func compose(list ...CommandVO) *cobra.Command {
	rootCmd := &cobra.Command{Use: os.Args[0]}
	for _, cvo := range list {
		c := cvo.CobraCommand
		c.Flags().BoolVarP(&vman.FlagVerbose, "verbose", "v", false, "Verbose debug output")
		var flags []string
		for _, ff := range cvo.FlagFunctions {
			flags = append(flags, ff(c))
		}
		for _, rf := range cvo.RequiredFlags {
			c.MarkFlagRequired(flags[rf])
		}
		rootCmd.AddCommand(c)
	}
	return rootCmd
}

//-------- flags definitions. All flags are static for app instance. --------//
func fFlagSet(c *cobra.Command) string {
	c.Flags().StringVarP(&flagSet, "set", "s", "", `Total number of IP including alias to create (if not exist) on sepecified Ethernet interface. 
If multiple ethernet interfaces match, then IP alias will be created for each.`)
	return "add"
}

func fFlagClean(c *cobra.Command) string {
	c.Flags().BoolVarP(&flagClean, "cleanup", "a", false, "Cleanup IP Alias on speciefied ethernet interface(s)")
	return "cleanup"
}

func fFlagEth(c *cobra.Command) string {
	c.Flags().StringVarP(&flagEth, "eth", "e", "", `Ethernet interfaces to create alias or start services on.
Wildcard allowed, but * should be escaped. ex: "eth\*" or "enp0s\*" 
If multiple ethernet interfaces match, service will start on all IP for match interface`)
	return "eth"
}

func fFlagIp(c *cobra.Command) string {
	c.Flags().StringVarP(&flagIp, "ip", "", "", `Ip address OR comma separated list of IPs
If --ip set, then --eth ignored`)
	return "ip"
}

func fFlagProcfile(c *cobra.Command) string {
	c.Flags().StringVarP(&flagProcfile, "procfile", "p", "", "Procfile to start services")
	return "procfile"
}

func fFlagBaseDir(c *cobra.Command) string {
	c.Flags().StringVarP(&flagBaseDir, "basedir", "d", ".", "Base dir for node")
	return "basedir"
}

func fFlagPort(c *cobra.Command) string {
	c.Flags().IntVarP(&flagPort, "port", "o", vman.DEF_RPC_PORT, "RPC Port to serve status etc")
	return "port"
}

func fFlagParent(c *cobra.Command) string {
	c.Flags().StringVarP(&flagParent, "parent", "r", "", "IP and Port of parent vipman to register childs IPs for proxy redirect and control")
	return "parent"
}

func fFlagChild(c *cobra.Command) string {
	c.Flags().StringVarP(&flagChild, "child", "l", "", "IP or NIC interfaces name of where vipman child started to configure proxy")
	return "child"
}

func fFlagProxy(c *cobra.Command) string {
	c.Flags().StringVarP(&flagProxy, "proxy", "x", "",
		`comma separated list of ports (listen:proxy) to redirect requests 
from parent (listen) to child (proxy) vipman(s):
If --parent flag set, then will register this vipman IP on parent vipman to proxy redirect.
If --parent flag not set, then will listen ports to redirect requests into registered IP.
On multiple child or IP registered parent proxy will use random pick from destination IP list.
`)
	return "proxy"
}

func fFlagProxyCmd(c *cobra.Command) string {
	c.Flags().StringVarP(&flagProxyCmd, "cmd", "c", "", "command (p)pause or (r)resume to control proxy")
	return "cmd"
}

func fFlagProcName(c *cobra.Command) string {
	c.Flags().StringVarP(&flagProcName, "proc", "", "", "name of process to restart as in Proc file")
	return "proc"
}

const CLIENT = `
This is a rpc call to running vipman service to perform the command. The command (except proxy) also will broadcasted to all underlying services `

var flagEth, flagProcfile, flagBaseDir, flagSet, flagIp string
var flagParent, flagChild, flagProxy, flagProxyCmd, flagProcName string
var flagClean bool
var flagPort int

func main() {
	compose(
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "prepare",
			Short: "Prepare Ethernet interfaces",
			Long: `Show list OR Create IP alias to network interfaces as preparetion to start services from proc file.
Create and Cleanup will require root priveledges, so run as "sudo vipman prepare --add"
If neither --add nor --clean flag set, then will show.
If either  --add and --clean flag set will fail.`,
			//			Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Prepare(flagEth, flagSet, flagClean) // done except clean
			},
		}, []func(*cobra.Command) string{fFlagEth, fFlagSet, fFlagClean}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "status",
			Short: "Short status of a vipman",
			Long: `Short status of a vipman processes including proxy and underlying vipmans.
Returns OK and exit code 0 if no errors` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCallNA("Status", flagPort)
			},
		}, []func(*cobra.Command) string{fFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "list",
			Short: "List processes",
			Long:  `Detailed list of processes uncluding underlying vipmans` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCallNA("List", flagPort)
			},
		}, []func(*cobra.Command) string{fFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "proxy",
			Short: "Configure proxy on parent to pause / resume childs vipman",
			Long: `Resume / Paused IP request to IP on Http ProxyRC.
If no --cmd flag, then list of proxy congurations
If no --child flags, then will list all child proxy configurations.
If --proxy set, then will override existing configuration` + CLIENT,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCall("Proxy", "", flagPort, &([]string{flagChild, flagProxyCmd, flagProxy, flagEth, flagIp}))
			},
		}, []func(*cobra.Command) string{fFlagPort, fFlagChild, fFlagProxyCmd, fFlagProxy, fFlagEth, fFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "start",
			Short: "Start Procfile on specified IP/eth",
			Long: `Process the Procfile (similar to goreman/foreman) but on IP alias and start set of processes.
If --ip flag set, then will start on particular IP(s), then --eth flag ignored,
If --eth flag set, then will start on all IPs alias on all matching interfaces.
`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Start(&vman.StartInfo{flagProcfile, flagPort, flagEth, flagIp, flagParent, flagBaseDir, flagProxy, make(map[string][]*vman.UIP), 0, ""})
			},
		}, []func(*cobra.Command) string{fFlagProcfile, fFlagPort, fFlagEth, fFlagIp, fFlagParent, fFlagBaseDir, fFlagProxy}, []int{0}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "stop",
			Short: "Stop process on all nodes or all processes on IP or interface",
			Long:  `Stop process by name or IP` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCall("Stop", "", flagPort, &[]string{flagIp, flagProcName})
			},
		}, []func(*cobra.Command) string{fFlagPort, fFlagIp, fFlagProcName}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "stop-all",
			Short: "Stop process on all nodes or all processes on IP or interface",
			Long:  `Stop all services on all vipman(s)` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCallNA("StopAll", flagPort)
			},
		}, []func(*cobra.Command) string{fFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "restart",
			Short: "Restart one process",
			Long:  `Restart processe by sending KILL signal and start again` + CLIENT,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCall("Restart", "", flagPort, &[]string{flagIp, flagProcName})
			},
		}, []func(*cobra.Command) string{fFlagPort, fFlagIp, fFlagProcName}, []int{1, 2}},
		/* /---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "restart-all",
			Short: "StopRC and start all with same set of settings",
			Long:  `Restart ` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RPCClientCall("RestartAll", "", flagPort, &[]string{flagEth, flagIp})
			},
		}, []func(*cobra.Command) string{fFlagPort, fFlagEth, fFlagIp}, []int{}},
		*/
	).Execute()
}
