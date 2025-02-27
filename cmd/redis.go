package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	clearKey "github.com/xiaozhuang-a/redisctl/pkg/service/redis"
)

func NewCmdRedisClean(ctx context.Context) *cobra.Command {
	clearCmd := clearKey.NewClearKey()
	cmd := &cobra.Command{
		Use:   "clean-key",
		Short: "clean key",
		Long:  ``,
		PreRun: func(cmd *cobra.Command, args []string) {
			logrus.Infof("input args: %v", args)
			if err := clearCmd.ParseParam(); err != nil {
				logrus.Fatal(err)
				return
			}
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("param: %+v", clearCmd.Param)
			clearCmd.Do(ctx)

			return nil
		},
	}
	param := clearCmd.Param

	cmd.Flags().StringVarP(&param.Host, "host", "H", "", "redis host")
	cmd.Flags().IntVarP(&param.Port, "port", "P", 6379, "redis port")
	cmd.Flags().StringVarP(&param.Username, "username", "u", "", "redis username")
	cmd.Flags().StringVarP(&param.Password, "password", "p", "", "redis password")
	cmd.Flags().IntVarP(&param.DB, "db", "d", 0, "redis db")
	cmd.Flags().BoolVarP(&param.IsPrefix, "prefix", "", false, "is prefix")
	cmd.Flags().StringSliceVarP(&param.Keys, "keys", "", []string{}, "keys")
	cmd.Flags().IntVarP(&param.DeleteBatch, "delete-batch", "", 300, "delete batch")
	cmd.Flags().IntVarP(&param.ScanBatch, "scan-batch", "", 300, "scan batch")
	cmd.Flags().IntVarP(&param.DeleteDelayMS, "delete-delay-ms", "", 100, "delete delay ms")
	cmd.Flags().BoolVarP(&param.DryRun, "dry-run", "", false, "dry run")
	cmd.Flags().BoolVarP(&param.OnlyNoExpire, "only-no-expire", "", false, "only no expire")
	cmd.Flags().BoolVarP(&param.OnlyHasExpire, "only-has-expire", "", false, "only has expire")
	cmd.Flags().IntVarP(&param.Concurrent, "concurrent", "", 10, "concurrent")
	cmd.Flags().StringVarP(&param.KeysPath, "keys-path", "", "", "keys path")
	cmd.Flags().BoolVarP(&param.EnableAliyunIScan, "enable-aliyun-iscan", "", false, "enable aliyun iscan")

	return cmd
}
