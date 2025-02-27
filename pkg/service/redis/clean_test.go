package redis

import (
	"context"
	"fmt"
	"github.com/xiaozhuang-a/redisctl/pkg/utils/engine"
	"strconv"
	"testing"
)

func TestClearKey_Do(t *testing.T) {
	cmd := NewClearKey()
	cmd.Param.Host = "r-0jld2f0fc58fc104.redis.rds.aliyuncs.com"
	cmd.Param.Port = 6379
	cmd.Param.Username = "root_dba"
	cmd.Param.Password = "3ST7eo2BYD4ixcQ!5d"
	cmd.Param.DB = 99
	cmd.Param.IsPrefix = true
	cmd.Param.Keys = []string{"xiaozhuang_test"}
	cmd.Param.OnlyNoExpire = true
	cmd.Param.Concurrent = 50
	cmd.Param.DeleteBatch = 1000
	cmd.Param.ScanBatch = 1000
	cmd.Param.DeleteDelayMS = 50
	cmd.Param.DryRun = true
	cmd.Do(context.Background())
}

func Test_DataBuilder(t *testing.T) {
	cli, err := engine.NewRedis(engine.Redis{
		Addr:     "10.11.24.104:6379",
		Password: "3ST7eo2BYD4ixcQ!5d",
		Username: "root",
		PoolSize: 1000,
		DB:       0,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5000; i++ {
		res := cli.Set(context.Background(), "xiaozhuang_test:"+strconv.Itoa(i), "value"+strconv.Itoa(i), 0)
		if res.Err() != nil {
			fmt.Println(res.Err())
			continue
		}
		fmt.Println(res.Val())
	}
}
