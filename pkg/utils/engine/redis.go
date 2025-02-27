package engine

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"

	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Redis struct {
	DB            int    `mapstructure:"db" json:"db" yaml:"db"`       // redis的哪个数据库
	Addr          string `mapstructure:"addr" json:"addr" yaml:"addr"` // 服务器地址:端口
	Username      string
	Password      string `mapstructure:"password" json:"password" yaml:"password"`                // 密码
	EnableTracing bool   `mapstructure:"enable-trace" json:"enableTrace" yaml:"enable-trace"`     // 是否会开启 Trace
	EnableTLS     bool   `mapstructure:"enable_tls" json:"enable_tls" yaml:"enable_tls"`          // 是否开启 tls
	MasterOnly    bool   `mapstructure:"master_only" json:"master_only" yaml:"master_only"`       // 是否只读主库，仅在集群模式下生效
	Protocol      int    `mapstructure:"protocol" json:"protocol" yaml:"protocol"`                // 协议版本,default 2
	PoolSize      int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`             // 连接池大小
	MinIdelConn   int    `mapstructure:"min_idel_conn" json:"min_idel_conn" yaml:"min_idel_conn"` // 最小空闲连接数
}

func NewRedis(redisCfg Redis) (client *redis.Client, err error) {
	options := &redis.Options{
		Addr:         redisCfg.Addr,
		Username:     redisCfg.Username,
		Password:     redisCfg.Password, // no password set
		DB:           redisCfg.DB,       // use default DB
		PoolSize:     redisCfg.PoolSize,
		MinIdleConns: redisCfg.MinIdelConn,
	}
	if redisCfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// 国内(腾讯)不支持3的协议，所以使用2的协议
	if redisCfg.Protocol == 0 {
		options.Protocol = 2
	}
	client = redis.NewClient(options)
	if err := injectRedisTracing(redisCfg.EnableTracing, client); err != nil {
		return nil, err
	}

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	} else {
		log.Infof("InitRedis success, cfg=%+v", redisCfg)
		return client, nil
	}
}

func InitClusterRedis(redisCfg Redis) (client *redis.ClusterClient, err error) {
	options := &redis.ClusterOptions{
		Addrs:        []string{redisCfg.Addr},
		Username:     redisCfg.Username,
		Password:     redisCfg.Password, // no password set
		PoolSize:     redisCfg.PoolSize,
		MinIdleConns: redisCfg.MinIdelConn,
	}
	// 国内(腾讯)不支持3的协议，所以使用2的协议
	if redisCfg.Protocol == 0 {
		options.Protocol = 2
	} else {
		options.Protocol = redisCfg.Protocol
	}
	if redisCfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if !redisCfg.MasterOnly {
		options.ReadOnly = true
		options.RouteRandomly = true
		log.Warnf("InitClusterRedis set ReadOnly and RouteRandomly")
	} else {
		log.Warnf("InitClusterRedis  MasterOnly")
	}
	client = redis.NewClusterClient(options)
	log.Info("InitClusterRedis NewClient done")

	if err = injectRedisTracing(redisCfg.EnableTracing, client); err != nil {
		return nil, err
	}

	_, err = client.Ping(context.Background()).Result()
	log.Info("InitClusterRedis Ping done")
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func injectRedisTracing(enableTracing bool, client redis.UniversalClient) error {
	if enableTracing {
		client.AddHook(RedisHook{})

		//return redisotel.InstrumentTracing(client)
	}
	return nil
}

type RedisHook struct{}

var _ redis.Hook = RedisHook{}

func (RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}
func (RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		now := time.Now()
		err := next(ctx, cmd)
		//utils.AddRedisCall(ctx, time.Now().Sub(now).Milliseconds(), err)
		dur := time.Now().Sub(now).Milliseconds()
		if dur > 100 {
			log.Infof("redis query slow, dur=%d, cmd=%s", dur, cmd.String())
		}

		return err
	}
}
func (RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		now := time.Now()
		err := next(ctx, cmds)
		dur := time.Now().Sub(now).Milliseconds()
		if dur > 100 {
			log.Infof("redis query slow, dur=%d, cmd=%v", dur, cmds)
		}

		return err
	}
}
