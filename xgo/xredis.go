package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/beego/beego/logs"
	"github.com/garyburd/redigo/redis"
)

type AbuRedisSubCallback func(string)
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
	db := GetConfigInt(fmt.Sprint(cfgname, ".db"), true, -1)
	password := GetConfigString(fmt.Sprint(cfgname, ".password"), true, "")
	maxidle := GetConfigInt(fmt.Sprint(cfgname, ".maxidle"), true, 0)
	maxactive := GetConfigInt(fmt.Sprint(cfgname, ".maxactive"), true, 0)
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

func (this *XRedis) getcallback(channel string) AbuRedisSubCallback {
	channel = fmt.Sprintf("%v:%v", project, channel)
	cb, ok := this.subscribecallbacks.Load(channel)
	if !ok {
		return nil
	}
	return cb.(AbuRedisSubCallback)
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
func (this *XRedis) Subscribe(channel string, callback AbuRedisSubCallback) {
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
	r := this.SetNx(key, "1", expire_second)
	return r
}

func (this *XRedis) ReleaseLock(key string) {
	this.Del(key)
}

func (this *XRedis) GetCacheMap(key string, cb func() *map[string]interface{}) *map[string]interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	data, _ := this.Get(key)
	if data != nil {
		jdata := map[string]interface{}{}
		json.Unmarshal(data, &jdata)
		return &jdata
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheArray(key string, cb func() *[]interface{}) *[]interface{} {
	key = fmt.Sprintf("%v:%v", project, key)
	data, _ := this.Get(key)
	if data != nil {
		jdata := []interface{}{}
		json.Unmarshal(data, &jdata)
		return &jdata
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheString(key string, cb func() string) string {
	key = fmt.Sprintf("%v:%v", project, key)
	data, _ := this.Get(key)
	if data != nil {
		return string(data)
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheInt(key string, cb func() int64) int64 {
	key = fmt.Sprintf("%v:%v", project, key)
	data, _ := this.Get(key)
	if data != nil {
		r, _ := strconv.ParseInt(string(data), 10, 64)
		return r
	} else {
		return cb()
	}
}

func (this *XRedis) GetCacheFloat(key string, cb func() float64) float64 {
	key = fmt.Sprintf("%v:%v", project, key)
	data, _ := this.Get(key)
	if data != nil {
		r, _ := strconv.ParseFloat(string(data), 32)
		return r
	} else {
		return cb()
	}
}
