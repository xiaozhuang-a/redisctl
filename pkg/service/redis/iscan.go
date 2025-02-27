package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
)

type IScanIteratorInterface interface {
	Next() bool
	Val() string
	Err() error
}

// IScanIterator 实现 ISCAN 命令的迭代器
type IScanIterator struct {
	ctx    context.Context
	rdb    *redis.Client
	nodeId int64
	match  string
	count  int64
	cursor uint64
	keys   []string
	index  int
	err    error

	done bool
}

// NewIScanIterator 创建一个新的 ISCAN 迭代器
func NewIScanIterator(ctx context.Context, rdb *redis.Client, nodeId int64, match string, count int64) IScanIteratorInterface {
	return &IScanIterator{
		ctx:    ctx,
		rdb:    rdb,
		nodeId: nodeId,
		match:  match,
		count:  count,
	}
}

// Next 获取下一个 key
func (it *IScanIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if it.index < len(it.keys) {
		it.index++
		return true
	}

	for {
		if it.done {
			return false
		}

		it.keys, it.cursor, it.err = it.iScan(it.ctx, it.cursor)
		if it.err != nil {
			return false
		}
		if it.cursor == 0 {
			it.done = true
		}

		if len(it.keys) > 0 {
			it.index = 1
			return true
		}
	}
}

func (it *IScanIterator) Val() string {
	if it.index == 0 || it.index > len(it.keys) || it.err != nil {
		return ""
	}
	return it.keys[it.index-1]
}

func (it *IScanIterator) Err() error {
	return it.err
}

func (it *IScanIterator) iScan(ctx context.Context, cursor uint64) ([]string, uint64, error) {
	res, err := it.rdb.Do(ctx, "iscan", it.nodeId, cursor, "match", it.match, "count", it.count).Result()
	if err != nil {
		return nil, 0, err
	}

	result, ok := res.([]interface{})
	if !ok || len(result) != 2 {
		return nil, 0, fmt.Errorf("unexpected result format: %v", res)
	}

	res1, ok := result[0].(string)
	if !ok {
		return nil, 0, fmt.Errorf("invalid cursor format: %v", result[0])
	}
	nextCursor, err := strconv.ParseUint(res1, 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("convert cursor to uint64 failed: %v", err)
	}

	keys, ok := result[1].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid keys format: %v", result[1])
	}

	strKeys := make([]string, len(keys))
	for i, key := range keys {
		strKeys[i] = key.(string)
	}

	return strKeys, nextCursor, nil
}
