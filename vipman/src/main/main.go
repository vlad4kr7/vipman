package main

import (
	"dbinvent.com/vipman/vman"
	"fmt"
	"github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	"os"
)

type cf struct {
	Ccom          *cobra.Command
	FlagFunctions []func(*cobra.Command)
}

var FlagVerbose bool

func compose(list ...cf) {
	rootCmd := &cobra.Command{Use: os.Args[0]}
	for _, cvo := range list {
		c := cvo.Ccom
		c.Flags().BoolVarP(&FlagVerbose, "verbose", "v", false, "Verbose debug output")
		for _, ff := range cvo.FlagFunctions {
			ff(c)
		}
		rootCmd.AddCommand(c)
	}
	rootCmd.Execute()
}

var FlagEth, FlagProcfile, FlagBaseDir, FlagPort string

func FFlagEth(c *cobra.Command) {
	c.Flags().StringVarP(&FlagEth, "eth", "e", "eth*", "Ethernet")
}

func FFlagProcfile(c *cobra.Command) {
	c.Flags().StringVarP(&FlagProcfile, "procfile", "p", "", "Procfile")
}

func FFlagBaseDir(c *cobra.Command) {
	c.Flags().StringVarP(&FlagBaseDir, "basedir", "d", "", "Base dir for node")
}

func FFlagPort(c *cobra.Command) {
	c.Flags().StringVarP(&FlagPort, "port", "", "", "App RPC Port to check status")
}

func main() {
	compose(
		cf{&cobra.Command{
			Use:   "prepare",
			Short: "Prepare Ethernet interfaces ",
			Long:  `List OR Create IP alias to network interfaces as preparetion to strart services from proc file`,
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				switch args[0] {
				case "add":
					vman.PrepareAdd(args[1:]...)
				default: // show
					vman.PrepareShow()
				}
			},
		}, []func(*cobra.Command){FFlagEth}},
		cf{&cobra.Command{
			Use:   "status",
			Short: "Short status of a vipman",
			Long: `Short status of a vipman processes including proxy and underlying vipmans.
Returns OK and exit code 0 if no errors
`,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcStatusCall()
			},
		}, []func(*cobra.Command){FFlagPort}},
		cf{&cobra.Command{
			Use:   "list",
			Short: "List processes",
			Long:  `Detailed list of processes uncluding underlying vipmans`,
			Run: func(cmd *cobra.Command, args []string) {
				vman.RpcListCall()
			},
		}, []func(*cobra.Command){FFlagPort}},
		cf{&cobra.Command{
			Use:   "join",
			Short: "Start a Procfile",
			Long:  `Start a Procfile and feed started processes to parent vipman`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagProcfile, FFlagPort}},
		cf{&cobra.Command{
			Use:   "pause",
			Short: "Pause apps on IP",
			Long:  `Stop dispatching requests to particular IP from Http Proxy`,
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort, FFlagEth}},
		cf{&cobra.Command{
			Use:   "resume",
			Short: "Resume paused IP",
			Long:  `Resume paused IP to dispatch to Http Proxy`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort, FFlagEth}},
		cf{&cobra.Command{
			Use:   "start",
			Short: "",
			Long:  `test test`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagProcfile, FFlagPort, FFlagEth}},
		cf{&cobra.Command{
			Use:   "stop",
			Short: "",
			Long:  `test test`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort, FFlagEth}},
		cf{&cobra.Command{
			Use:   "stop-all",
			Short: "",
			Long:  `test test`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort}},
		cf{&cobra.Command{
			Use:   "restart",
			Short: "",
			Long:  `test test`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort, FFlagEth}},
		cf{&cobra.Command{
			Use:   "restart-all",
			Short: "",
			Long:  `test test`,
			//Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("test v?%t \n", FlagVerbose)
			},
		}, []func(*cobra.Command){FFlagPort}},
	)
}
