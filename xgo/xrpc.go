package xgo

import (
	"errors"
	"fmt"
	"net/rpc"
	"time"

	"github.com/beego/beego/logs"
	"github.com/spf13/viper"
)

type XRpc struct {
	host   string
	client *rpc.Client
}

func XRpcInit(configpath string) *XRpc {
	xrpc := &XRpc{}
	return xrpc.Init(configpath)
}

func (this *XRpc) Init(configpath string) *XRpc {
	this.host = viper.GetString(configpath)
	if this.host == "" {
		return this
	}
	for {
		client, err := rpc.Dial("tcp", this.host)
		if err != nil {
			logs.Error("rpc建立链接失败:", this.host)
			time.Sleep(time.Second)
			continue
		}
		this.client = client
		break
	}
	return this
}

func (this *XRpc) Call(name string, data H) (*XMap, error) {
	if this.client == nil {
		logs.Debug("rpc not connected", this.host)
		return nil, errors.New("rpc not connected")
	}
	result := XMap{}
	err := this.client.Call(name, XMap{RawData: data}, &result)
	if err != nil {
		logs.Error(fmt.Sprintf("XRpc %s Error:", name), err)
		return &result, err
	}
	return &result, nil
}
