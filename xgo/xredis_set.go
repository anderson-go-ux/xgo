package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/beego/beego/logs"
)

func (this *XRedis) SAdd(key string, value interface{}) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	typename := reflect.TypeOf(value).Name()
	strval := ""
	switch typename {
	case "int", "int32", "int64", "float", "float32", "float64", "string":
		strval = fmt.Sprint(value)
	default:
		bytes, _ := json.Marshal(&value)
		strval = string(bytes)
	}
	_, err := conn.Do("sadd", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return nil
}

func (this *XRedis) SMembers(key string) []interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("smembers", key)
	if err != nil {
		logs.Error(err.Error())
		return []interface{}{}
	}
	if ret == nil {
		return []interface{}{}
	}
	arr := ret.([]interface{})
	return arr
}

func (this *XRedis) SRem(key string, value interface{}) []interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	data := []interface{}{}
	data = append(data, key)
	typename := reflect.TypeOf(value).Name()
	strval := ""
	switch typename {
	case "int", "int32", "int64", "float", "float32", "float64", "string":
		strval = fmt.Sprint(value)
	default:
		bytes, _ := json.Marshal(&value)
		strval = string(bytes)
	}
	data = append(data, strval)
	_, err := conn.Do("srem", data...)
	if err != nil {
		logs.Error(err.Error())
		return []interface{}{}
	}
	return nil
}

func (this *XRedis) SIsMember(key string, value interface{}) (bool, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	typename := reflect.TypeOf(value).Name()
	strval := ""
	switch typename {
	case "int", "int32", "int64", "float", "float32", "float64", "string":
		strval = fmt.Sprint(value)
	default:
		bytes, _ := json.Marshal(&value)
		strval = string(bytes)
	}
	ret, err := conn.Do("smembers", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return false, err
	}
	return ret.(int64) == 1, nil
}

func (this *XRedis) SCard(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("scard", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	return ret.(int64), nil
}
