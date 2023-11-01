package main

import (
	"github.com/zhms/xgo/xgo"
)

var db *xgo.XDb = &xgo.XDb{}
var redis *xgo.XRedis = &xgo.XRedis{}
var http *xgo.XHttp = &xgo.XHttp{}

func main() {
	xgo.Init()
	db.Init("server.db")
	redis.Init("server.redis")
	http.Init("server.http", redis)
	http.InitWs("/sapi/ws")
	http.InitWs("/capi/ws")
	xgo.AdminInit(http, db, redis)
	xgo.BackupDb(db, "db.sql")
	xgo.Run()
}
