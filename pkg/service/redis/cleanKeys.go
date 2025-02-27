package redis

import (
	"context"
	"fmt"
	"github.com/juju/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/xiaozhuang-a/redisctl/pkg/utils/engine"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ClearKey struct {
	ctx  context.Context
	logg *logrus.Entry

	Param *ClearKeyParam

	// 用于缓存清理队列，非阻塞队列
	queue chan string

	// 用于检查key ttl属性的队列
	ttlQueue chan string

	// 用于结束
	done chan struct{}

	// 用于delay delete
	deleteCh chan struct{}
	// 原子锁，int类型
	deleteCount atomic.Int64
}

type ClearKeyParam struct {
	DryRun   bool
	Host     string
	Port     int
	Username string
	Password string
	DB       int

	IsPrefix bool
	Keys     []string
	KeysPath string
	//Backup    bool
	//BackupDir string

	DeleteBatch   int
	ScanBatch     int
	DeleteDelayMS int
	// 只清理没有过期时间的key
	OnlyNoExpire bool
	// 只清理有过期时间的key
	OnlyHasExpire bool

	Concurrent int

	//logLevel string
	EnableAliyunIScan bool
}

var ErrHostEmpty = errors.New("host is empty")

func NewClearKey() *ClearKey {
	return &ClearKey{
		logg:  logrus.WithField("cmd", "clean-key"),
		Param: &ClearKeyParam{},
		queue: make(chan string, 500000),
		done:  make(chan struct{}),
		//deleteCh: make(chan struct{}, 1),
		ttlQueue: make(chan string, 500000),
	}
}

func (c *ClearKey) ParseParam() error {
	if c.Param.Host == "" {
		return ErrHostEmpty
	}
	if c.Param.Port == 0 {
		c.Param.Port = 6379
	}
	if c.Param.Password == "" {
		return errors.New("password is empty")
	}
	if len(c.Param.Keys) == 0 && c.Param.KeysPath == "" {
		return errors.New("keys is empty")
	}
	// 禁止*
	for _, key := range c.Param.Keys {
		if key == "*" {
			return errors.New("keys can not contains *")
		}
	}
	if c.Param.DeleteBatch == 0 {
		c.Param.DeleteBatch = 500
	}
	if c.Param.ScanBatch == 0 {
		c.Param.ScanBatch = 1000
	}
	if c.Param.DeleteDelayMS == 0 {
		c.Param.DeleteDelayMS = 50
	}
	if c.Param.OnlyHasExpire && c.Param.OnlyNoExpire {
		return errors.New("only-has-expire and only-no-expire can not be true at the same time")
	}

	if c.Param.Concurrent == 0 {
		c.Param.Concurrent = 10
	}

	// 获取keysPath的keys，并合并到keys
	if c.Param.KeysPath != "" {
		keys, err := engine.ReadLinesStrings(c.Param.KeysPath)
		if err != nil {
			return errors.Trace(err)
		}
		c.Param.Keys = append(c.Param.Keys, keys...)
	}

	c.Param.Keys = c.validateKey(c.Param.Keys)

	// 初始化deleteCh，如果并发大于1，使用缓冲channel
	if c.Param.Concurrent > 1 {
		c.deleteCh = make(chan struct{}, c.Param.Concurrent)
	} else {
		c.deleteCh = make(chan struct{}, 1)
	}

	return nil
}

func (c *ClearKey) Do(ctx context.Context) {
	c.ctx = ctx
	c.run()
}

func (c *ClearKey) run() {
	//if err := c.validateParam(); err != nil {
	//	c.logg.Panicf("validate param error: %v", err)
	//}
	// 定时打印status
	go func() {
		// 打印deletecount变化速率
		internal := 5
		ticker := time.NewTicker(time.Second * time.Duration(internal))
		defer ticker.Stop()
		var lastCount int64
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-c.done:
				count := c.deleteCount.Load()
				c.logg.Infof("delete total count: %d", count)
				return
			case <-ticker.C:
				count := c.deleteCount.Load()
				if c.Param.OnlyHasExpire || c.Param.OnlyNoExpire {
					c.logg.Infof("delete count: %d, speed: %d/s,queue: %d,ttlQueue: %d", count, (count-lastCount)/int64(internal), len(c.queue), len(c.ttlQueue))
				} else {
					c.logg.Infof("delete count: %d, speed: %d/s,queue: %d", count, (count-lastCount)/int64(internal), len(c.queue))
				}
				lastCount = count
			}
		}
	}()
	if c.Param.OnlyNoExpire || c.Param.OnlyHasExpire {
		go c.ttlCheck()
	}
	go c.scan()
	go c.delete()
	for {
		select {
		case <-c.ctx.Done():
			c.logg.Warnf("clean key execution canceled")
			return
		case <-c.done:
			c.logg.Infof("clean key execution completed")
			time.Sleep(time.Second * 1)
			return
		}
	}
}

func (c *ClearKey) scan() {
	cli := c.newRedisClient()
	defer cli.Close()

	for _, key := range c.Param.Keys {
		if key == "" {
			continue
		}
		if c.Param.IsPrefix {
			c.scanOneKey(cli, key)
		} else {
			c.queue <- key
		}
	}
	c.logg.Infof("scan key completed")
	if c.Param.OnlyHasExpire || c.Param.OnlyNoExpire {
		c.closeTTLQueue()
	} else {
		c.closeQueue()
	}
}

func (c *ClearKey) matchKey(key string) string {
	if c.Param.IsPrefix {
		return fmt.Sprintf("%s*", key)
	}
	return key
}

func (c *ClearKey) scanOneKey(cli *redis.Client, key string) {
	match := c.matchKey(key)
	llo := c.logg.WithField("scan match ", match)

	// 并发对所有节点同时使用iscan
	if c.Param.EnableAliyunIScan {
		llo.Infof("scan key use aliyun iscan")
		wg := sync.WaitGroup{}
		var nodeId int
		for {
			exist, err := ExistNode(c.ctx, cli, int64(nodeId))
			if err != nil {
				llo.Errorf("exist node %d error: %v", nodeId, err)
				return
			}
			if !exist {
				break
			}

			wg.Add(1)
			go func(nodeId int) {
				defer wg.Done()
				iter := NewIScanIterator(c.ctx, cli, int64(nodeId), match, int64(c.Param.ScanBatch))
				for iter.Next() {
					if iter.Val() == "" {
						llo.Warn("iscan key empty")
						continue
					}
					if c.Param.OnlyNoExpire || c.Param.OnlyHasExpire {
						c.ttlQueue <- iter.Val()
					} else {
						c.queue <- iter.Val()
					}
				}
				if err := iter.Err(); err != nil {
					llo.Errorf("iscan key error: %v", err)
					return
				}
			}(nodeId)
			nodeId++
		}
		wg.Wait()

	} else {
		iter := cli.Scan(c.ctx, 0, match, int64(c.Param.ScanBatch)).Iterator()
		for iter.Next(c.ctx) {
			if iter.Val() == "" {
				llo.Warn("scan key empty")
				continue
			}
			if c.Param.OnlyNoExpire || c.Param.OnlyHasExpire {
				c.ttlQueue <- iter.Val()
			} else {
				c.queue <- iter.Val()
			}
		}
		if err := iter.Err(); err != nil {
			llo.Errorf("scan key error: %v", err)
			return
		}
	}

	llo.Infof("scan key completed")
}

func (c *ClearKey) closeQueue() {
	close(c.queue)
}

func (c *ClearKey) closeTTLQueue() {
	close(c.ttlQueue)
}

// 并发检查key的ttl属性
func (c *ClearKey) ttlCheck() {
	cli := c.newRedisClient()
	defer cli.Close()
	defer c.closeQueue()

	maxConcurrent := make(chan struct{}, c.Param.Concurrent)
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-c.ctx.Done():
			c.logg.Infof("ttl check canceled")
			return
		case key, ok := <-c.ttlQueue:
			maxConcurrent <- struct{}{}
			wg.Add(1)
			go func(key string) {
				defer func() {
					<-maxConcurrent
					wg.Done()
				}()
				if key != "" {
					ttl := cli.TTL(c.ctx, key).Val()
					if c.Param.OnlyNoExpire && ttl == time.Duration(-1) {
						c.queue <- key
					}
					if c.Param.OnlyHasExpire && ttl != time.Duration(-1) {
						c.queue <- key
					}
				}
			}(key)

			if !ok {
				c.logg.Infof("ttl check completed")
				return
			}
		}
	}
}

func (c *ClearKey) newRedisClient() *redis.Client {
	cli, err := engine.NewRedis(engine.Redis{
		Addr:        c.Param.Host + ":" + strconv.Itoa(c.Param.Port),
		Password:    c.Param.Password,
		DB:          c.Param.DB,
		Username:    c.Param.Username,
		MinIdelConn: 5,
		PoolSize:    1000,
	})
	if err != nil {
		c.logg.Panicf("new redis client error: %v", err)
	}
	return cli
}

func (c *ClearKey) delete() {
	// 处理延迟delete
	for i := 0; i < c.Param.Concurrent; i++ {
		go func() {
			for {
				select {
				case <-c.ctx.Done():
					c.logg.Infof("delete delay process canceled")
					return
				case c.deleteCh <- struct{}{}:
					time.Sleep(time.Millisecond * time.Duration(c.Param.DeleteDelayMS))
				}
			}
		}()
	}

	cli := c.newRedisClient()
	defer func() {
		c.done <- struct{}{}
	}()
	defer cli.Close()

	var over bool

	batchTimeout := time.Millisecond * 100
	tt := time.NewTimer(batchTimeout)
	var batchKeys []string

	wg := sync.WaitGroup{}
	deleteCh := make(chan struct{}, c.Param.Concurrent)
	for !over {
		tt.Reset(batchTimeout)

		select {
		case <-c.ctx.Done():
			c.logg.Infof("delete key canceled")
			return
		case key, ok := <-c.queue:
			if key != "" {
				batchKeys = append(batchKeys, key)
			}
			if !ok {
				over = true
				break
			}
			if len(batchKeys) < c.Param.DeleteBatch {
				continue
			}
		case <-tt.C:
			//c.logg.Infof("get queue key timeout,current queue len: %d", len(c.queue))
		}
		if len(batchKeys) > 0 {
			wg.Add(1)
			deleteCh <- struct{}{}
			go func(keys []string) {
				defer func() {
					<-deleteCh
					wg.Done()
				}()
				c.deleteKeys(cli, keys)
			}(batchKeys)
			batchKeys = []string{}
		}
	}
	wg.Wait()

	c.logg.Infof("delete key completed")
}

// 单进程批量删就足够，因为scan是单进程的，目前无法利用上多分片
func (c *ClearKey) deleteKeys(cli *redis.Client, keys []string) {
	if len(keys) == 0 {
		return
	}
	select {
	case <-c.ctx.Done():
		c.logg.Infof("deleteKeys canceled")
		return
	case <-c.deleteCh:
		if c.Param.DryRun {
			c.logg.Infof("dry run delete keys: %v", keys)
		} else {
			if err := cli.Unlink(c.ctx, keys...).Err(); err != nil {
				c.logg.Errorf("delete key %s error: %v", keys, err)
			} else {
				c.deleteCount.Add(int64(len(keys)))
			}
		}
	}
}

func ExistNode(ctx context.Context, rdb *redis.Client, nodeId int64) (bool, error) {
	_, err := rdb.Do(ctx, "iinfo", nodeId).Result()
	if err != nil {
		if err.Error() == "ERR no such db node" {
			return false, nil
		}
		return false, errors.Trace(err)
	}

	return true, nil
}

// 校验key 并去掉前后空格和换行符、*
func (c *ClearKey) validateKey(key []string) []string {
	var keys []string
	for _, k := range key {
		k = strings.Trim(k, " ")
		k = strings.Trim(k, "\r")
		k = strings.Trim(k, "\n")
		k = strings.Trim(k, "*")
		if k == "" {
			continue
		}
		keys = append(keys, k)
	}
	c.logg.Infof("validate key: %v", keys)
	return keys
}
