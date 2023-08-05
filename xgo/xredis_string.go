package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/beego/beego/logs"
)

func (this *XRedis) Set(key string, value interface{}, expireseconds int) error {
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
	if expireseconds > 0 {
		_, err := conn.Do("setex", key, expireseconds, strval)
		if err != nil {
			logs.Error(err.Error())
			return err
		}
		return nil
	} else {
		_, err := conn.Do("set", key, strval)
		if err != nil {
			logs.Error(err.Error())
			return err
		}
		return nil
	}
}

func (this *XRedis) Get(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("get", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) GetRange(key string, start int, end int) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("getrange", key, start, end)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) GetSet(key string, value interface{}) ([]byte, error) {
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
	ret, err := conn.Do("getset", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	fmt.Println(string(ret.([]byte)))
	return ret.([]byte), nil
}

func (this *XRedis) SetNx(key string, value interface{}) bool {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	key = fmt.Sprintf("%v:%v", project, key)
	typename := reflect.TypeOf(value).Name()
	strval := ""
	switch typename {
	case "int", "int32", "int64", "float", "float32", "float64", "string":
		strval = fmt.Sprint(value)
	default:
		bytes, _ := json.Marshal(&value)
		strval = string(bytes)
	}
	r, err := conn.Do("setnx", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	ir := r.(int64)
	return ir == 1
}

func (this *XRedis) SetRange(key string, offset int, value interface{}) error {
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
	_, err := conn.Do("setrange", key, offset, strval)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) StrLen(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("strlen", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	return ret.(int64), nil
}

func (this *XRedis) IncrByFloat(key string, value float64) (float64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("incrbyfloat", key, value)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	r, _ := strconv.ParseFloat(string(ret.([]byte)), 32)
	return r, nil
}

func (this *XRedis) Append(key string, value interface{}) error {
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
	_, err := conn.Do("append", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}
