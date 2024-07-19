package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

type CmdInterface interface {
	Do(ctx context.Context)
	ParseParam() error
}

var Command = &cobra.Command{
	Use: "redisctl",
	//Version:          version.Get().GitVersion,
	Short:            "",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.AddCommand(NewCmdRedisClean(context.Background()))
}
