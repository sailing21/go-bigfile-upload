package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type redisOp struct {
	redisPool *redis.Pool
}

/**
检查是否已满
*/
func (rop *redisOp) chunkIsFull(fileMd5 string) int {
	cnt, e2 := redis.Int(rop.redisPool.Get().Do("ZCARD", fileMd5))
	errorHanler(e2)
	return cnt
}
func (rop *redisOp) getMem(fileMd5 string) []string {
	res, e := redis.Strings(rop.redisPool.Get().Do("ZRANGE", fileMd5, 0, -1)) // 无需分数 （顺序）
	errorHanler(e)
	return res
}
func (rop *redisOp) getMemWithScore(fileMd5 string) map[string]string {
	res, e := redis.StringMap(rop.redisPool.Get().Do("ZRANGE", fileMd5, 0, -1, "WITHSCORES")) // 无需分数 （顺序）
	errorHanler(e)
	return res
}
func (rop *redisOp) clearSet(fileMd5 string) {
	_, e := rop.redisPool.Get().Do("DEL", fileMd5)
	errorHanler(e)
}

/**
添加块
*/
func (rop *redisOp) chunkAdd(idx int, fileMd5 string, pieceMd5 string) {
	rop.redisPool.Get().Do("ZADD", fileMd5, idx, pieceMd5)
}

/**
文件信息存储、查询
*/
func (rop *redisOp) fileInfo(fileMd5 string, filename string, filepath string, filetype string) string {
	res, e := redis.String(rop.redisPool.Get().Do("HMSET", "hashmap_"+fileMd5,
		"filename", filename,
		"filepath", filepath,
		"filetype", filetype,
	))
	fmt.Println(res)
	errorHanler(e)
	return res
}

func (rop *redisOp) getFileinfo(fileMd5 string) map[string]string {
	res, e := redis.StringMap(rop.redisPool.Get().Do("HGETALL", "hashmap_"+fileMd5))
	errorHanler(e)
	return res
}

/**
合并加锁
*/
func (rop *redisOp) merging(fileMd5 string) int {
	res, e := redis.Int(rop.redisPool.Get().Do("SADD", "merging_list", fileMd5))
	errorHanler(e)
	return res
}

/**
是否正在合并
返回int
*/
func (rop *redisOp) isMerging(fileMd5 string) int {
	res, e := redis.Int(rop.redisPool.Get().Do("SISMEMBER", "merging_list", fileMd5))
	errorHanler(e)
	return res
}

/**
合并完删除
*/
func (rop *redisOp) delMerging(fileMd5 string) int {
	res, e := redis.Int(rop.redisPool.Get().Do("SREM", "merging_list", fileMd5))
	errorHanler(e)
	return res
}

//生成一个连接池对象
func (rop *redisOp) newPool() *redis.Pool {
	server := ":6379"
	return &redis.Pool{
		MaxIdle:     10000,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			//if _, err := c.Do("AUTH", ""); err != nil {
			//	c.Close()
			//	return nil, err
			//}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
func (rop *redisOp) init() *redisOp {
	rop.redisPool = rop.newPool()
	return rop
}
