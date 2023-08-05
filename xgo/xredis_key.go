package xgo

import (
	"fmt"

	"github.com/beego/beego/logs"
)

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
