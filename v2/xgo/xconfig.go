package xgo

import (
	"fmt"

	"github.com/beego/beego/logs"
	"github.com/spf13/viper"
)

// 获取yaml配置int64
// key 配置名 eg: server.http.port
// require 是否必须有该项
// invalval 如果必须有该项,该项的无效值 例如 必须配置server.http.port,改项值不可为0 GetConfigInt("server.http.port",true,0)
func GetConfigInt(key string, require bool, invalval int64) int64 {
	val := viper.GetInt64(key)
	if require && val == invalval {
		err := fmt.Sprint("read config error:", key)
		logs.Error(err)
		panic(err)
	}
	return val
}

// 获取yaml配置string
// key 配置名 eg: server.http.host
// require 是否必须有该项
// invalval 如果必须有该项,该项的无效值 例如 必须配置server.http.port,改项值不可为"" GetConfigString("server.http.host",true,"")
func GetConfigString(key string, require bool, invalval string) string {
	val := viper.GetString(key)
	if require && val == invalval {
		err := fmt.Sprint("read config error:", key)
		logs.Error(err)
		panic(err)
	}
	return val
}

// 获取yaml配置float64
// key 配置名 eg: server.http.timeout
// require 是否必须有该项
// invalval 如果必须有该项,该项的无效值 例如 必须配置server.http.timeout,改项值不可为 1.0 GetConfigFloat("server.http.timeout",true,1.0)
func GetConfigFloat(key string, require bool, invalval float64) float64 {
	val := viper.GetFloat64(key)
	if require && val == invalval {
		err := fmt.Sprint("read config error:", key)
		logs.Error(err)
		panic(err)
	}
	return val
}
