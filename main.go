package main

import (
	"database/sql"
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
	db.Transaction(func(tx *sql.Tx) error {
		x, err := db.Table("x_config").Tx(tx).Count()
		if err != nil {
			return err
		}
		fmt.Println(x)
		return nil
	})
	xgo.Run()
}
