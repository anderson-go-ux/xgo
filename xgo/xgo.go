package xgo

import (
	"time"

	mrand "math/rand"

	"github.com/beego/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var env string = "2"
var project string
var module string

/*
go get github.com/beego/beego/logs
go get github.com/spf13/viper
go get github.com/gin-gonic/gin
go get github.com/go-redis/redis
go get github.com/garyburd/redigo/redis
go get github.com/go-sql-driver/mysql
go get github.com/satori/go.uuid
go get github.com/gorilla/websocket
go get github.com/jinzhu/gorm
go get github.com/imroc/req
go get github.com/go-playground/validator/v10
go get github.com/go-playground/universal-translator
go get code.google.com/p/mahonia
go get github.com/360EntSecGroup-Skylar/excelize
go clean -modcache
*/
func Init() {
	mrand.NewSource(time.Now().UnixNano())
	gin.SetMode(gin.ReleaseMode)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterFile, `{"filename":"_log/logfile.log","maxsize":10485760}`)
	logs.SetLogger(logs.AdapterConsole, `{"color":true}`)
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigName("configx")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		logs.Error("读取配置文件失败", err)
		return
	}
}
