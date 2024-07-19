package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xiaozhuang-a/redisctl/cmd"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

var Cmd = &cobra.Command{
	Use: "redisctl",
	//Version:          version.Get().GitVersion,
	Short:            "",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	cmd.Init()
	Cmd.AddCommand(cmd.RedisCmd)
}

func main() {
	err := Cmd.Execute()
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
}
