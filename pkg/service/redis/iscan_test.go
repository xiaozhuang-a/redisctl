package redis

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/xiaozhuang-a/redisctl/pkg/utils/engine"
	"sync"
	"sync/atomic"
	"testing"
)

func TestIScanIterator_Iterator(t *testing.T) {
	cli, err := engine.NewRedis(engine.Redis{
		Addr:        "127.0.0.1:6379",
		Password:    "123456",
		DB:          0,
		MinIdelConn: 5,
	})
	if err != nil {
		panic(err)
	}

	// 测试 IScanIterator和scan结果是否一致
	wg := sync.WaitGroup{}
	var nodeId int
	var count atomic.Int64
	count.Store(0)
	for {
		exist, err := ExistNode(context.TODO(), cli, int64(nodeId))
		if err != nil {
			logrus.Errorf("exist node %d error: %v", nodeId, err)
			return
		}
		if !exist {
			break
		}

		wg.Add(1)
		go func(nodeId int) {
			defer wg.Done()
			it := NewIScanIterator(context.Background(), cli, int64(nodeId), "*", 1000)
			for it.Next() {
				if it.Val() == "" {
					logrus.Warnf("empty key")
					continue
				}

				//logrus.Infof("key: %s", it.Val())
				count.Add(1)
			}
			if err := it.Err(); err != nil {
				panic(err)
				return
			}
		}(nodeId)
		nodeId++
	}
	wg.Wait()

	iter := cli.Scan(context.Background(), 0, "*", 5000).Iterator()
	var count2 int
	for iter.Next(context.Background()) {
		if iter.Val() == "" {
			logrus.Warnf("empty key")
			continue
		}

		//logrus.Infof("key: %s", iter.Val())
		count2++
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}
	logrus.Infof("iscan count: %d, scan count: %d", count.Load(), count2)
}
