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
func FFlagAdd(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagAdd, "add", "a", "", `Number of IP alias to create on sepecified Ethernet interface. 
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
	c.Flags().StringVarP(&vman.FlagIp, "ip", "", "", `Ip address OR comma separeted list of IPs`)
	return "ip"
}

func FFlagProCommandVOile(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagProcfile, "proCommandVOile", "p", "", "ProCommandVOile to start services")
	return "proCommandVOile"
}

func FFlagBaseDir(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagBaseDir, "basedir", "d", "", "Base dir for node")
	return "basedir"
}

func FFlagPort(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagPort, "port", "o", "", "RPC Port to serve status etc")
	return "port"
}

func FFlagParent(c *cobra.Command) string {
	c.Flags().StringVarP(&vman.FlagParent, "parent", "r", "", "IP and Port of parent vipman to register")
	return "parent"
}

const CLIENT = `
This is a rpc call to running vipman service to perform the command. The command also will broadcasted to all underlying services`

func main() {
	compose(
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "prepare",
			Short: "Prepare Ethernet interfaces",
			Long: `Show list OR Create IP alias to network interfaces as preparetion to strart services from proc file.
Create and Cleanup will require root priveledges, so run as "sudo vipman prepare --add"
If neither --add nor --clean flag set, then will show.
If either  --add and --clean flag set will fail.`,
			//			Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Prepare() // done except clean
			},
		}, []func(*cobra.Command) string{FFlagEth, FFlagAdd, FFlagClean}, []int{0}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "status",
			Short: "Short status of a vipman",
			Long: `Short status of a vipman processes including proxy and underlying vipmans.
Returns OK and exit code 0 if no errors` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcStatusCall() //todo
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "list",
			Short: "List processes",
			Long:  `Detailed list of processes uncluding underlying vipmans` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcListCall() //todo
			},
		}, []func(*cobra.Command) string{FFlagPort}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "join",
			Short: "Start a ProCommandVOile. Same as start, but join the root vipman",
			Long: `Start a ProCommandVOile and feed started processes to parent vipman. 
Useful to run set of services on VMWare / Virtual box.` + CLIENT,
			Run: func(cmd *cobra.Command, args []string) {
				vman.Join()
			},
		}, []func(*cobra.Command) string{FFlagProCommandVOile, FFlagEth, FFlagParent, FFlagPort}, []int{0, 1, 2}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "pause",
			Short: "Pause apps on IP or all IP on interface (if set)",
			Long: `Stop dispatching requests to particular IP from Http Proxy. 
Will fail if neither --ip or --eth  flag set` + CLIENT,
			//			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Pause()
			},
		}, []func(*cobra.Command) string{FFlagEth, FFlagPort, FFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "resume",
			Short: "Resume paused IP",
			Long: `Resume paused IP or all IP on interface (if set) to dispatch to Http Proxy.
Will resume all paused if  neither --ip or --eth  flag set.` + CLIENT,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Resume()
			},
		}, []func(*cobra.Command) string{FFlagPort, FFlagEth, FFlagIp}, []int{}},
		//---------------------------------------------------------------------------//
		CommandVO{&cobra.Command{
			Use:   "start",
			Short: "Start ProCommandVOile on specified IP/eth",
			Long: `Process the ProCommandVOile (similar to goreman/foreman) but on IP alias and start set of processes.
If --eth flag set, then will start on all IPs alias on all matching interfaces.
If --ip flag set, then will start on particular IP(s)`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				vman.Stop()
			},
		}, []func(*cobra.Command) string{FFlagProCommandVOile, FFlagEth, FFlagIp, FFlagPort}, []int{0}},
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
