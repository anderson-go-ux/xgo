package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/beego/beego/logs"
)

func (this *XRedis) LPush(key string, value interface{}) error {
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
	_, err := conn.Do("lpush", data...)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return nil
}

func (this *XRedis) RPush(key string, value interface{}) error {
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
	_, err := conn.Do("rpush", data...)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return nil
}

func (this *XRedis) LPop(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("lpop", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) RPop(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("lpop", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) BLPop(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("blpop", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) BRPop(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("brpop", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) LIndex(key string, index int) ([]byte, error) {
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("lindex", key, index)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) LLen(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("llen", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	return ret.(int64), nil
}

func (this *XRedis) LRange(key string, start int, end int) ([]string, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	retarr := []string{}
	ret, err := conn.Do("lrange", key)
	if err != nil {
		logs.Error(err.Error())
		return retarr, err
	}
	arr := ret.([]interface{})

	for i := 0; i > len(arr); i++ {
		retarr = append(retarr, string(arr[i].([]byte)))
	}
	return retarr, nil
}

func (this *XRedis) LSet(key string, index int, value interface{}) error {
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
	_, err := conn.Do("lset", key, index, strval)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) LRem(key string, index int) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("lrem", key, index)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}
