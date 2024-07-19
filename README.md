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
Usage:                                                                                   ✔ ▓▒░
  redisctl [command]

Available Commands:
  clean-key   clean key
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -h, --help   help for redisctl

```

## clean-key
```shell
Usage:
  redisctl clean-key [flags]

Flags:
      --concurrent int        concurrent (default 10)
  -d, --db int                redis db
      --delete-batch int      delete batch (default 500)
      --delete-delay-ms int   delete delay ms (default 80)
      --dry-run               dry run
  -h, --help                  help for clean-key
  -H, --host string           redis host
      --keys strings          keys
      --only-has-expire       only has expire
      --only-no-expire        only no expire
  -p, --password string       redis password
  -P, --port int              redis port (default 6379)
      --prefix                is prefix
      --scan-batch int        scan batch (default 500)
  -u, --username string       redis username
```

**How to control the load on the target instance？**
+ `--concurrent` control the ttl query concurrent,Effective only when --only-has-expire or --only-no-expire exists
+ `--delete-delay-ms` control the delay time between delete operation，This is the most effective way

**Common parameters**
+ --only-has-expire, only delete keys that have expiration time
+ --only-no-expire, only delete keys that have no expiration time
+ --keys, specify the key to delete,eg: --keys "quest,test22"
+ --prefix, whether to use the key as a prefix to delete


## Example

```shell
./redisctl clean-key --host  -p  --keys "quest" --prefix true
```

