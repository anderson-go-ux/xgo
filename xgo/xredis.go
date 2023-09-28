package xgo

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/beego/beego/logs"
	"github.com/garyburd/redigo/redis"
)

type XRedisSubCallback func(string)
type XRedis struct {
	redispool          *redis.Pool
	pubconnection      *redis.PubSubConn
	host               string
	port               int
	db                 int
	password           string
	recving            bool
	subscribecallbacks sync.Map
}

func (this *XRedis) Init(cfgname string) {
	if this.redispool != nil {
		return
	}
	host := GetConfigString(fmt.Sprint(cfgname, ".host"), true, "")
	port := GetConfigInt(fmt.Sprint(cfgname, ".port"), true, 0)
	db := int(GetConfigInt(fmt.Sprint(cfgname, ".db"), true, -1))
	password := GetConfigString(fmt.Sprint(cfgname, ".password"), true, "")
	maxidle := int(GetConfigInt(fmt.Sprint(cfgname, ".maxidle"), true, 0))
	maxactive := int(GetConfigInt(fmt.Sprint(cfgname, ".maxactive"), true, 0))
	idletimeout := GetConfigInt(fmt.Sprint(cfgname, ".idletimeout"), true, 0)
	this.redispool = &redis.Pool{
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: time.Duration(idletimeout) * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", fmt.Sprint(host, ":", port),
				redis.DialPassword(password),
				redis.DialDatabase(db),
			)
			if err != nil {
				logs.Error(err)
				panic(err)
			}
			return con, nil
		},
	}
	conn, err := redis.Dial("tcp", fmt.Sprint(host, ":", port),
		redis.DialPassword(password),
		redis.DialDatabase(db),
	)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	this.pubconnection = new(redis.PubSubConn)
	this.pubconnection.Conn = conn
	this.recving = false
	logs.Debug("连接redis 成功:", host, port, db)
}

func (this *XRedis) getcallback(channel string) XRedisSubCallback {
	channel = fmt.Sprintf("%v:%v", project, channel)
	cb, ok := this.subscribecallbacks.Load(channel)
	if !ok {
		return nil
	}
	return cb.(XRedisSubCallback)
}

func (this *XRedis) subscribe(channels ...string) {
	this.pubconnection.Subscribe(redis.Args{}.AddFlat(channels)...)
	if !this.recving {
		this.recving = true
		go func() {
			for {
				imsg := this.pubconnection.Receive()
				msgtype := reflect.TypeOf(imsg).Name()
				if msgtype == "Message" {
					msg := imsg.(redis.Message)
					callback := this.getcallback(msg.Channel)
					if callback != nil {
						callback(string(msg.Data))
					}
				}
			}
		}()
	}
}

// ////////////////////////////////////////////////////////////////////////////////
func (this *XRedis) Subscribe(channel string, callback XRedisSubCallback) {
	channel = fmt.Sprintf("%v:%v", project, channel)
	this.subscribecallbacks.Store(channel, callback)
	this.subscribe(channel)
}

func (this *XRedis) Publish(key string, value interface{}) error {
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
	_, err := conn.Do("publish", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) GetLock(key string, expire_second int) bool {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	if expire_second <= 0 {
		r, err := conn.Do("setnx", key, "1")
		if err != nil {
			logs.Error(err.Error())
			return false
		}
		ir := r.(int64)
		if ir == 1 && expire_second > 0 {
			conn.Do("expire", key, expire_second)
		}
		return ir == 1
	} else {
		r, err := conn.Do("set", key, "1", "EX", expire_second, "NX")
		if err != nil {
			logs.Error(err.Error())
			return false
		}
		ir := r.(int64)
		return ir == 1
	}
}

func (this *XRedis) ReleaseLock(key string) {
	this.Del(key)
}

func (this *XRedis) GetCacheMap(key string, cb func() (*XMap, error)) (*XMap, error) {
	data, _ := this.Get(key)
	if data != nil {
		jdata := map[string]interface{}{}
		json.Unmarshal(data, &jdata)
		xmap := XMap{}
		xmap.RawData = jdata
		return &xmap, nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheMaps(key string, cb func() (*XMaps, error)) (*XMaps, error) {
	data, _ := this.Get(key)
	if data != nil {
		jdata := []map[string]interface{}{}
		json.Unmarshal(data, &jdata)
		xmaps := XMaps{}
		for i := 0; i < len(jdata); i++ {
			xmaps.RawData = append(xmaps.RawData, XMap{RawData: jdata[i]})
		}
		return &xmaps, nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheString(key string, cb func() (string, error)) (string, error) {
	data, _ := this.Get(key)
	if data != nil {
		return string(data), nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheStrings(key string, cb func() (*[]string, error)) (*[]string, error) {
	data, _ := this.Get(key)
	if data != nil {
		jdata := []string{}
		json.Unmarshal(data, &jdata)
		return &jdata, nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheInt(key string, cb func() (int64, error)) (int64, error) {
	data, _ := this.Get(key)
	if data != nil {
		return ToInt(data), nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheInts(key string, cb func() (*[]int64, error)) (*[]int64, error) {
	data, _ := this.Get(key)
	if data != nil {
		jdata := []int64{}
		json.Unmarshal(data, &jdata)
		return &jdata, nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheFloat(key string, cb func() (float64, error)) (float64, error) {
	data, _ := this.Get(key)
	if data != nil {
		return ToFloat(data), nil
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheFloats(key string, cb func() (*[]float64, error)) (*[]float64, error) {
	data, _ := this.Get(key)
	if data != nil {
		jdata := []float64{}
		json.Unmarshal(data, &jdata)
		return &jdata, nil
	} else {
		return cb()
	}
}

func (this *XRedis) Ping() error {
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("ping")
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) Del(key string) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("del", key)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) Dump(key string) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("dump", key)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return ret.([]byte), nil
}

func (this *XRedis) Exists(key string) (bool, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("exists", key)
	if err != nil {
		logs.Error(err.Error())
		return false, err
	}
	return ret.(int64) == 1, nil
}

func (this *XRedis) Expire(key string, second_expire int) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("expire", key, second_expire)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) ExpireAt(key string, second_timestamp int64) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("expireat", key, second_timestamp)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) PExpire(key string, millisecond_expire int) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("pexpire", key, millisecond_expire)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) PExpireAt(key string, millisecond_timestamp int) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("pexpireat", key, millisecond_timestamp)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) Keys(key string) ([]string, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("keys", key)
	if err != nil {
		logs.Error(err.Error())
		return []string{}, err
	}
	retarr := []string{}
	arr := ret.([]interface{})
	for i := 0; i < len(arr); i++ {
		retarr = append(retarr, string(arr[i].([]byte)))
	}
	return retarr, nil
}

func (this *XRedis) Move(key string, db int) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("move", key, db)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) Persist(key string) error {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("persist", key)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) PTtl(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("pttl", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	return ret.(int64), nil
}

func (this *XRedis) Ttl(key string) (int64, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("ttl", key)
	if err != nil {
		logs.Error(err.Error())
		return 0, err
	}
	return ret.(int64), nil
}

func (this *XRedis) RandomKey() (string, error) {
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("randomkey")
	if err != nil {
		logs.Error(err.Error())
		return "", err
	}
	fmt.Println(string(ret.([]byte)))
	return string(ret.([]byte)), nil
}

func (this *XRedis) Rename(oldkey string, newkey string) error {
	oldkey = fmt.Sprintf("%v:%v", project, oldkey)
	newkey = fmt.Sprintf("%v:%v", project, newkey)
	conn := this.redispool.Get()
	defer conn.Close()
	_, err := conn.Do("rename", oldkey, newkey)
	if err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

func (this *XRedis) RenameNx(oldkey string, newkey string) (bool, error) {
	oldkey = fmt.Sprintf("%v:%v", project, oldkey)
	newkey = fmt.Sprintf("%v:%v", project, newkey)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("renamenx", oldkey, newkey)
	if err != nil {
		logs.Error(err.Error())
		return false, err
	}
	return ret.(int64) == 1, nil
}

func (this *XRedis) Type(key string) (string, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("type", key)
	if err != nil {
		logs.Error(err.Error())
		return "", err
	}
	return ret.(string), nil
}

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

func (this *XRedis) BLPop(key string, timeout int) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("blpop", key, timeout)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	if ret != nil {
		arr := ret.([]interface{})
		return arr[1].([]byte), nil
	}
	return nil, errors.New("失败")
}

func (this *XRedis) BRPop(key string, timeout int) ([]byte, error) {
	key = fmt.Sprintf("%v:%v", project, key)
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do("brpop", key, timeout)
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
	if ret == nil {
		return nil, nil
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

func (this *XRedis) SetNx(key string, value interface{}, expire_second int) bool {
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
	r, err := conn.Do("setnx", key, strval)
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	ir := r.(int64)
	if ir == 1 && expire_second > 0 {
		conn.Do("expire", key, expire_second)
	}
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

func (this *XRedis) Do(commond string, args ...interface{}) (interface{}, error) {
	conn := this.redispool.Get()
	defer conn.Close()
	ret, err := conn.Do(commond, args...)
	return ret, err
}

func (this *XRedis) LoadScript(name string, script string) error {
	conn := this.redispool.Get()
	defer conn.Close()
	hasher := sha1.New()
	hasher.Write([]byte(script))
	hashHex := hex.EncodeToString(hasher.Sum(nil))
	key := fmt.Sprintf("%v:__redis_script:%v", project, name)
	ret, err := conn.Do("script", "exists", hashHex)
	if err != nil {
		return err
	}
	retarr := ret.([]interface{})
	if retarr[0].(int64) == 1 {
		_, err = conn.Do("set", key, hashHex)
		if err != nil {
			return err
		}
		return nil
	}
	ret, err = conn.Do("script", "load", script)
	if err != nil {
		return err
	}
	_, err = conn.Do("set", key, hashHex)
	if err != nil {
		return err
	}
	return nil
}

func (this *XRedis) ScriptEval(name string, args ...interface{}) (interface{}, error) {
	conn := this.redispool.Get()
	defer conn.Close()
	key := fmt.Sprintf("%v:__redis_script:%v", project, name)
	ret, err := conn.Do("get", key)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, errors.New(fmt.Sprintf("获取脚本哈希失败:%v", name))
	}
	scripthash := string(ret.([]byte))
	params := []interface{}{}
	params = append(params, scripthash)
	params = append(params, len(args))
	params = append(params, args...)
	ret, err = conn.Do("evalsha", params...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
