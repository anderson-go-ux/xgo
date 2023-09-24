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
	//xgo.BackupDb(db, "db.sql")
	//opt := `[{"field":"Id","name":"Id","values":{}},{"field":"SellerId","name":"运营商","values":{"1":"初始运营商"}},{"field":"ChannelId","name":"渠道","values":{"1":"初始渠道"}},{"field":"UserId","name":"UserId","values":{}},{"field":"State","name":"状态","values":{"1":"启用","2":"禁用"}},{"field":"Account","name":"账号","values":{}},{"field":"Token","name":"最后登录token","values":{}},{"field":"NickName","name":"昵称","values":{}},{"field":"PhoneNum","name":"电话号码","values":{}},{"field":"Email","name":"Email地址","values":{}},{"field":"TopAgent","name":"顶级代理","values":{}},{"field":"Agents","name":"代理","values":{}},{"field":"Agent","name":"上级代理","values":{}},{"field":"CreateTime","name":"创建时间","values":{},"date":1}]`
	//opt = ""
	//db.Table("x_user").OrderBy("id desc").Export("user1", opt)
	x, _ := redis.GetCacheInts("test", func() (*[]int64, error) {
		data := []int64{1, 2, 3}
		redis.Set("test", data, 0)
		return &data, nil
	})
	fmt.Println("fffffff", x)
	xgo.Run()
}
