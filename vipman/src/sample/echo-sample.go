package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var FlagIp string
var FlagHertbeat, FlagPort int

func echoHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok, "+c.Request().RequestURI)
}

func hertbeat() {
	for {
		log.Println("bip")
		time.Sleep(time.Duration(FlagHertbeat) * time.Second)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Run: func(cmd *cobra.Command, args []string) {
			if FlagHertbeat > 0 {
				go hertbeat()
			}
			port := strconv.Itoa(FlagPort)
			log.Println("Starting echo web service on port: " + port)

			e := echo.New()
			e.Use(middleware.Logger())
			e.Use(middleware.Recover())
			e.GET("/*", echoHandler)

			e.Logger.Fatal(e.Start(FlagIp + ":" + port))
		},
	}
	rootCmd.Flags().IntVarP(&FlagHertbeat, "hertbeat", "b", 0, "Hertbeat interval. Default set to zero (0) to disable hertbeat")
	rootCmd.Flags().IntVarP(&FlagPort, "port", "p", 8080, "Echo server port")
	rootCmd.Flags().StringVarP(&FlagIp, "ip", "i", "0.0.0.0", "Echo server binding IP")
	rootCmd.Execute()

}
