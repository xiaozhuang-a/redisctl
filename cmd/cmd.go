package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

type CmdInterface interface {
	Do(ctx context.Context)
	ParseParam() error
}

var RedisCmd = &cobra.Command{
	Use:              "redis",
	Aliases:          []string{"redis"},
	Short:            "tools for redis command",
	Long:             ``,
	TraverseChildren: true,
}

func Init() {
	RedisCmd.AddCommand(NewCmdRedisClean(context.Background()))
}
