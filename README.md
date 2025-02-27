# redisctl
redis tools ,e.g clean redis keys

# Usage
## build
```bash
go mod tidy
go build -o redisctl
```

show commands

```shell
Usage:
  redisctl redis [command]

Aliases:
  redis, redis

Available Commands:
  clean-key   clean key

Flags:
  -h, --help   help for redis


```

## clean-key
```shell
Usage:
  redisctl redis clean-key [flags]

Flags:
      --concurrent int        concurrent (default 10)
  -d, --db int                redis db
      --delete-batch int      delete batch (default 300)
      --delete-delay-ms int   delete delay ms (default 100)
      --dry-run               dry run
      --enable-aliyun-iscan   enable aliyun iscan
  -h, --help                  help for clean-key
  -H, --host string           redis host
      --keys strings          keys
      --keys-path string      keys path
      --only-has-expire       only has expire
      --only-no-expire        only no expire
  -p, --password string       redis password
  -P, --port int              redis port (default 6379)
      --prefix                is prefix
      --scan-batch int        scan batch (default 300)
  -u, --username string       redis username
```

**How to control the load on the target instance？**
+ `--concurrent` control the ttl query concurrent,effective only when --only-has-expire or --only-no-expire exists
+ `--delete-delay-ms` control the delay time between delete operation，This is the most effective way

**Common parameters**
+ --only-has-expire, only delete keys that have expiration time
+ --only-no-expire, only delete keys that have no expiration time
+ --keys, specify the key to delete,eg: --keys "quest,test22"
+ --prefix, whether to use the key as a prefix to delete


## Example

```shell
./redisctl clean-key --host 127.0.0.1 -p '123'  --keys "quest" --prefix true

./redisctl clean-key --host 127.0.0.1 -p '123'  --keys-path /tmp/keys.txt --prefix true --dry-run

./redisctl redis clean-key --host 127.0.0.1 -p '123' --prefix true  --enable-aliyun-iscan  --keys 'tj-4kffnjohzy:'  --scan-batch 2000 --delete-delay-ms 0 --delete-batch 3000 --concurrent 20 --dry-run
```

