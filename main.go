package main

import (
	"fmt"

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
	t1 := db.Table("x_user").Select("Id").Where("Id < ?", 100)
	t2 := db.Table("x_config").Select("Id").Where("Id > ?", 225)
	data, _ := db.Union(t1, t2)
	fmt.Println(data.Maps())
	xgo.Run()
}
