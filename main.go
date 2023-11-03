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
	xgo.ShowTable(db, "x_config")
	xgo.Run()
}

type XConfig struct {
	Id          int    `gorm:"column:Id" json:"Id"`                   //
	SellerId    int    `gorm:"column:SellerId" json:"SellerId"`       // 运营商
	ChannelId   int    `gorm:"column:ChannelId" json:"ChannelId"`     // 渠道
	ConfigName  string `gorm:"column:ConfigName" json:"ConfigName"`   // 配置名
	ConfigValue string `gorm:"column:ConfigValue" json:"ConfigValue"` // 配置值
	ForClient   int    `gorm:"column:ForClient" json:"ForClient"`     // 该配置客户端是否能获取
	Memo        string `gorm:"column:Memo" json:"Memo"`               // 备注
	CreateTime  string `gorm:"column:CreateTime" json:"CreateTime"`   // 创建时间
}

func (this *XConfig) TableName() string {
	return "x_config"
}
