package main

import (
	"github.com/zhms/xgo/xgo"
)

var db *xgo.XDb = &xgo.XDb{}
var redis *xgo.XRedis = &xgo.XRedis{}

func main() {
	xgo.Init()
	db.Init("server.db")
	redis.Init("server.redis")

	redis.IncrByFloat("testa", -1.2)

	xgo.Run()
}
