package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

var flagIP string
var flagHertbeat, flagPort int

func echoHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok, "+c.Request().RequestURI+"\n")
}

func heartbeat() {
	for {
		log.Println("bip")
		time.Sleep(time.Duration(flagHertbeat) * time.Second)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Run: func(cmd *cobra.Command, args []string) {
			if flagHertbeat > 0 {
				go heartbeat()
			}
			port := strconv.Itoa(flagPort)
			log.Println("Starting echo web service on " + flagIP + ":" + port)

			e := echo.New()
			e.Use(middleware.Logger())
			e.Use(middleware.Recover())
			e.GET("/*", echoHandler)

			e.Logger.Fatal(e.Start(flagIP + ":" + port))
		},
	}
	rootCmd.Flags().IntVarP(&flagHertbeat, "heartbeat", "b", 0, "Hertbeat interval. Default set to zero (0) to disable heartbeat")
	rootCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Echo server port")
	rootCmd.Flags().StringVarP(&flagIP, "ip", "i", "0.0.0.0", "Echo server binding IP")
	rootCmd.Execute()

}
