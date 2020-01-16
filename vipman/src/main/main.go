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

//-------- Flags definitions. All flags are static for app instance. --------//
func FFlagSet(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagSet, "set", "s", "", `Total number of IP including alias to create (if not exist) on sepecified Ethernet interface. 
If multiple ethernet interfaces match, then IP alias will be created for each.`)
	return "add"
}
func FFlagClean(c *cobra.Command) string {
	c.Flags().BoolVarP(&vman.FlagClean, "cleanup", "c", false, "Cleanup IP Alias on speciefied ethernet interface(s)")
	return "cleanup"
}
func FFlagEth(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagEth, "eth", "e", "", `Ethernet interfaces to create alias or start services on.
Wildcard allowed, but should be escaped. ex: "eth\*" or "enp0s\*" 
If multiple ethernet interfaces match, service will start on all IP alias for will be created for each.`)
	return "eth"
}
func FFlagIp(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagIp, "ip", "", "", `Ip address OR comma separated list of IPs`)
	return "ip"
}

func FFlagProcfile(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagProcfile, "procfile", "p", "", "Procfile to start services")
	return "procfile"
}

func FFlagBaseDir(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagBaseDir, "basedir", "d", "", "Base dir for node")
	return "basedir"
}

func FFlagPort(c *cobra.Command) string {
	c.Flags().IntVarP(&vman.FlagPort, "port", "o", vman.DEF_RPC_PORT, "RPC Port to serve status etc")
	return "port"
}

func FFlagParent(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagParent, "parent", "r", "", "IP and Port of parent vipman to register")
	return "parent"
}

func FFlagChild(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagChild, "child", "l", "", "IP or NIC interfaces name of where vipman child started to configure proxy")
	return "child"
}

func FFlagProxy(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagProxy, "proxy", "x", "", "command (p)pause or (r)resume to control proxy")
	return "proxy"
}

const CLIENT = `
This is a rpc call to running vipman service to perform the command. The command (except proxy) also will broadcasted to all underlying services `

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
				vman.Prepare() // done except clean
			},
		}, []func(*cobra.Command) string{FFlagEth, FFlagSet, FFlagClean}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "status",
			Short: "Short status of a vipman",
			Long: `Short status of a vipman processes including proxy and underlying vipmans.
Returns OK and exit code 0 if no errors` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcStatusCall()
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "list",
			Short: "List processes",
			Long:  `Detailed list of processes uncluding underlying vipmans` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcListCall()
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "proxy",
			Short: "Configure proxy on parent to pause / resume childs vipman",
			Long: `Resume / Paused IP request to IP on Http Proxy.
If no --child flags, then will list all child proxy configurations.` + CLIENT,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Proxy()
			},
		}, []func(*cobra.Command) string{FFlagChild, FFlagProxy, FFlagPort, FFlagEth, FFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "start",
			Short: "Start ProCommandVOile on specified IP/eth",
			Long: `Process the ProCommandVOile (similar to goreman/foreman) but on IP alias and start set of processes.
If --ip flag set, then will start on particular IP(s), then --eth flag ignored,
If --eth flag set, then will start on all IPs alias on all matching interfaces.`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Start()
			},
		}, []func(*cobra.Command) string{FFlagProcfile, FFlagEth, FFlagIp, FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "stop",
			Short: "Stop process on all nodes or all processes on IP or interface",
			Long: `Stop process on all nodes:
 - like stop-all: this is a default action if no additional flag set
OR All processes on IP:
 - set --ip flag 
OR All  processes on interface:
 - set --eth flag
` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.Stop()
			},
		}, []func(*cobra.Command) string{FFlagPort, FFlagEth, FFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "stop-all",
			Short: "Alias for stop",
			Long:  `Stop all services on all vipman(s)` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.StopAll()
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "restart",
			Short: "Restart something",
			Long: `Restart all processes by sending KILL signal and start again 
 - like restart-all: this is a default action if no additional flag set
OR All processes on IP:
 - set --ip flag 
OR All  processes on interface:
 - set --eth flag` + CLIENT,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Restart()
			},
		}, []func(*cobra.Command) string{FFlagPort, FFlagEth, FFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "restart-all",
			Short: "Stop and start all with same set of settings",
			Long:  `Restart ` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RestartAll()
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
	).Execute()
}
