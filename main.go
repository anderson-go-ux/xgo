package main

import (
	"github.com/zhms/xgo/xgo"
)

var db *xgo.XDb = &xgo.XDb{}
var redis *xgo.XRedis = &xgo.XRedis{}
var http *xgo.XHttp = &xgo.XHttp{}

type XConfig struct {
	Id          int    `gorm:"column:Id"`
	SellerId    int    `gorm:"column:SellerId"`
	ChannelId   int    `gorm:"column:ChannelId"`
	ConfigName  string `gorm:"column:ConfigName"`
	ConfigValue string `gorm:"column:ConfigValue"`
	EditAble    int    `gorm:"column:EditAble"`
	ShowAble    int    `gorm:"column:ShowAble"`
	ForClient   int    `gorm:"column:ForClient"`
	CreateTime  string `gorm:"column:CreateTime"`
}

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
	xgo.Run()
}
