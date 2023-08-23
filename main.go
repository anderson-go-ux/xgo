package main

import (
	"fmt"

	"github.com/zhms/xgo/xgo"
)

var db *xgo.XDb = &xgo.XDb{}
var redis *xgo.XRedis = &xgo.XRedis{}
var http *xgo.XHttp = &xgo.XHttp{}

var fullatuh = `
{
	"系统首页": { "查" : 1},
	"系统管理": {
		"运营商管理": { "查": 0,"增": 0,"删": 0,"改": 0},
		"系统设置":   { "查": 1,"增": 1,"删": 1,"改": 1},
		"渠道管理":   { "查": 1,"增": 1,"删": 1,"改": 1},
		"账号管理":   { "查": 1,"增": 1,"删": 1,"改": 1},
		"角色管理":   { "查": 1,"增": 1,"删": 1,"改": 1},
		"登录日志":   { "查": 1},
		"操作日志":   { "查": 1}
	}
}
`

func main() {
	xgo.Init()
	db.Init("server.db")
	redis.Init("server.redis")
	http.Init("server.http", redis)
	http.InitWs("/api/ws")
	xgo.AdminInit(http, db, redis, fullatuh)
	xgo.BackupDb(db, "db.sql")
	xf, _ := db.Table("x_user_pool").Insert(xgo.H{
		"UserId": 1,
		"Ip":     "abc",
	})
	f, _ := (*xf).LastInsertId()
	fmt.Println(f)
	xgo.Run()
}
