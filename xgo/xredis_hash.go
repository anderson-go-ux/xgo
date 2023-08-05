package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/beego/beego/logs"
)

func (this *XRedis) HDel(key string, field string) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("hdel", key, field)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return nil
}

func (this *XRedis) HExists(key string, field string) (bool, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("hexists", key, field)
	if err != nil {
		logs.Error(err.Error())
		return false, err
	}
	return ret.(int64) == 1, nil
}

func (this *XRedis) HSet(key string, field string, value interface{}) error {
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
	_, err := conn.Do("hset", key, field, strval)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) HGet(key string, field string) interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("hget", key, field)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return ret
}

func (this *XRedis) HGetAll(key string) *map[string]interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("hgetall", key)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	arrret := ret.([]interface{})
	if len(arrret) == 0 {
		return nil
	}
	mapret := map[string]interface{}{}
	for i := 0; i < len(arrret); i++ {
		if i%2 == 0 {
			mapret[string(arrret[i].([]byte))] = string(arrret[i+1].([]byte))
		}
	}
	return &mapret
}

func (this *XRedis) HIncrByFloat(key string, field string, value float64) (float64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("hincrbyfloat", key, field, value)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	r, _ := strconv.ParseFloat(string(ret.([]byte)), 32)
	return r, nil
}

func (this *XRedis) HKeys(key string) ([]string, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	keys, err := conn.Do("hkeys", key)
	ikeys := keys.([]interface{})
	strkeys := []string{}
	if err != nil {
		logs.Error(err.Error())
		return strkeys, err
	}
	for i := 0; i < len(ikeys); i++ {
		strkeys = append(strkeys, string(ikeys[i].([]byte)))
	}
	return strkeys, nil
}

func (this *XRedis) HLen(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("hlen", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, nil
	}
	return ret.(int64), nil
}

func (this *XRedis) HSetNx(key string, field string, value interface{}) (bool, error) {
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
	ret, err := conn.Do("hsetnx", key, field, strval)
	fmt.Println(ret)
	if err != nil {
		logs.Error(err.Error())
		return false, err
	}
	return ret.(int64) == 1, nil
}

func (this *XRedis) HMGet(key string, fields ...interface{}) *map[string]interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	data := []interface{}{}
	data = append(data, key)
	data = append(data, fields...)
	ret, err := conn.Do("hmget", data...)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	arrret := ret.([]interface{})
	if len(arrret) == 0 {
		return nil
	}
	mapret := map[string]interface{}{}
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if arrret[i] != nil {
			mapret[field.(string)] = string(arrret[i].([]byte))
		}
	}
	return &mapret
}
