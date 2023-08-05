package xgo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/beego/beego/logs"
	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
)

/*
错误码:
	0. 成功
	1. 没有配置token的redis
	2. 请求header未填写x-token值
	3. 未登录或者登录过期了
	4. 参数格式错误,参数必须是json格式
	5. 权限不足
	6. 参数校验错误
	7. 重复请求
*/

type DBLogCallback func(interface{})

type XHttp struct {
	gin   *gin.Engine
	token *XRedis

	upgrader             websocket.Upgrader
	request_log_callback DBLogCallback
}

type XHttpContent struct {
	gin       *gin.Context
	TokenData string
	Token     string
	reqdata   string
}

type HttpResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type XHttpHandler func(*XHttpContent)

func (this *XHttpContent) RequestData(obj interface{}) error {
	err := json.Unmarshal([]byte(this.reqdata), &obj)
	if err != nil {
		this.RespErr(6, err.Error())
		return err
	}
	validator := val.New()
	err = validator.Struct(obj)
	if err != nil {
		this.RespErr(6, err.Error())
		return err
	}
	return nil
}

func (this *XHttpContent) Query(key string) string {
	return this.gin.Query(key)
}

func (this *XHttpContent) GetIp() string {
	return this.gin.ClientIP()
}

func (this *XHttpContent) Gin() *gin.Context {
	return this.gin
}

func (c *XHttpContent) Host() string {
	return strings.Split(c.gin.Request.Host, ":")[0]
}

func (ctx *XHttpContent) Put(key string, value interface{}) {
	if ctx.gin.Keys == nil {
		ctx.gin.Keys = make(map[string]interface{})
	}
	if ctx.gin.Keys["REPONSE_DATA"] == nil {
		ctx.gin.Keys["REPONSE_DATA"] = make(map[string]interface{})
	}
	if len(key) <= 0 || key == "" {
		ctx.gin.Keys["REPONSE_DATA"] = value
		return
	}
	ctx.gin.Keys["REPONSE_DATA"].(map[string]interface{})[key] = value
}

func (ctx *XHttpContent) RespOK(objects ...interface{}) {
	resp := new(HttpResponse)
	resp.Code = 0
	resp.Msg = "success"
	if len(objects) > 0 {
		ctx.Put("", objects[0])
	}
	resp.Data = ctx.gin.Keys["REPONSE_DATA"]
	if resp.Data == nil {
		resp.Data = make(map[string]interface{})
	}
	ctx.gin.JSON(http.StatusOK, resp)
}

func (ctx *XHttpContent) RespJson(obj any) {
	ctx.gin.JSON(http.StatusOK, obj)
}

func (ctx *XHttpContent) RespErr(data ...interface{}) {
	resp := new(HttpResponse)
	if len(data) == 2 {
		resp.Code = data[0].(int)
		resp.Msg = data[1].(string)
	} else {
		resp.Msg = data[0].(string)
		resp.Code = -1
	}
	resp.Data = ctx.gin.Keys["REPONSE_DATA"]
	ctx.gin.JSON(http.StatusOK, resp)
}

func abuhttpcors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type, x-token, Content-Length, X-Requested-With")
		context.Header("Access-Control-Allow-Methods", "GET,POST")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}

func (this *XHttp) Init(cfgname string, token *XRedis) {
	this.token = token
	port := GetConfigInt(cfgname+".port", true, 0)
	this.gin = gin.New()
	this.gin.Use(abuhttpcors())
	go func() {
		bind := fmt.Sprint("0.0.0.0:", port)
		this.gin.Run(bind)
	}()
	this.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	if this.token != nil {
		go func() {
			for {
				logdata, _ := this.token.BLPop(fmt.Sprintf("%s:token:http_request", project), 86400)
				if logdata != nil && this.request_log_callback != nil {
					this.request_log_callback(logdata)
				}
			}
		}()
	}
	logs.Debug("http listen:", port)
}

func (this *XHttp) Static(relativePaths string, root string) {
	this.gin.Static(relativePaths, root)
}

func (this *XHttp) OnPost(path string, handler XHttpHandler) {
	this.OnPostWithAuth(path, handler, "", false)
}

func (this *XHttp) OnPostWithAuth(path string, handler XHttpHandler, auth string, googleverify bool) {
	this.gin.POST(path, func(gc *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				logs.Error(err)
				stack := debug.Stack()
				logs.Error(string(stack))
			}
		}()
		body, _ := io.ReadAll(gc.Request.Body)
		strbody := string(body)
		if len(strbody) == 0 {
			strbody = "{}"
		}
		ctx := &XHttpContent{gc, "", "", strbody}
		if this.token == nil {
			ctx.RespErr(1, "未配置token")
			return
		}
		tokenstr := gc.GetHeader("x-token")
		if len(tokenstr) == 0 {
			ctx.RespErr(2, "请在header填写:x-token")
			return
		}
		rediskey := fmt.Sprintf("%v:token:%v", project, tokenstr)
		tokendata, _ := this.token.Get(rediskey)
		if tokendata == nil {
			ctx.RespErr(3, "未登录或登录已过期")
			return
		}
		keystr := fmt.Sprintf("post%v%v", gc.Request.URL.Path, tokenstr)
		reqid := Md5(keystr)
		lockkey := fmt.Sprintf("%v:token:lock:%v", project, reqid)
		if !this.token.SetNx(lockkey, "1", 10) {
			ctx.RespErr(7, "请勿重复请求")
			return
		}
		defer func() {
			this.token.Del(lockkey)
		}()
		this.token.Expire(rediskey, 7*86400)
		ctx.TokenData = string(tokendata)
		ctx.Token = tokenstr
		var iauthdata interface{}
		jbody := map[string]interface{}{}
		err := json.Unmarshal([]byte(strbody), &jbody)
		if err != nil {
			ctx.RespErr(4, "参数必须是json格式")
			return
		}
		jtoken := map[string]interface{}{}
		json.Unmarshal([]byte(ctx.TokenData), &jtoken)
		iauthdata = jtoken["AuthData"]
		if googleverify {
			gc, ok := jbody["GoogleCode"]
			if !ok {
				ctx.RespErr(8, "请填写谷歌验证码")
				return
			}
			gcstr := InterfaceToString(gc)
			if len(gcstr) == 0 {
				ctx.RespErr(8, "请填写谷歌验证码")
				return
			}
			if strings.Index(env, "prd") > 0 {
				gsstr := GetMapString(&jtoken, "GoogleSecret")
				if !VerifyGoogleCode(gsstr, gcstr) {
					ctx.RespErr(9, "谷歌验证码不正确")
					return
				}
			}
		}
		jlog := H{"ReqPath": gc.Request.URL.Path,
			"ReqData": jbody, "Account": jtoken["Account"], "UserId": jtoken["UserId"],
			"SellerId": jtoken["SellerId"], "ChannelId": jtoken["ChannelId"], "Ip": ctx.GetIp(), "Token": tokenstr}
		strlog, _ := json.Marshal(&jlog)
		this.token.RPush(fmt.Sprintf("%s:token:http_request", project), string(strlog))
		if len(auth) > 0 {
			spauth := strings.Split(auth, ".")
			m := spauth[0]
			s := spauth[1]
			o := spauth[2]
			if len(spauth) == 3 && iauthdata != nil {
				authdata := make(map[string]interface{})
				json.Unmarshal([]byte(iauthdata.(string)), &authdata)
				im, imok := authdata[m]
				if !imok {
					ctx.RespErr(5, "权限不足")
					return
				}
				is, isok := im.(map[string]interface{})[s]
				if !isok {
					ctx.RespErr(5, "权限不足")
					return
				}
				io, iook := is.(map[string]interface{})[o]
				if !iook {
					ctx.RespErr(5, "权限不足")
					return
				}
				if strings.Index(reflect.TypeOf(io).Name(), "float64") < 0 {
					ctx.RespErr(5, "权限不足")
					return
				}
				if InterfaceToInt(io) != 1 {
					ctx.RespErr(5, "权限不足")
					return
				}
			}
		}
		handler(ctx)
	})
}

func (this *XHttp) OnPostNoAuth(path string, handler XHttpHandler) {
	this.gin.POST(path, func(gc *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				logs.Error(err)
				stack := debug.Stack()
				logs.Error(string(stack))
			}
		}()
		body, _ := io.ReadAll(gc.Request.Body)
		strbody := string(body)
		if len(strbody) == 0 {
			strbody = "{}"
		}
		ctx := &XHttpContent{gc, "", "", strbody}
		if this.token != nil {
			jbody := map[string]interface{}{}
			err := json.Unmarshal([]byte(strbody), &jbody)
			if err != nil {
				ctx.RespErr(4, "参数必须是json格式")
				return
			}
			jlog := gin.H{"ReqPath": gc.Request.URL.Path, "ReqData": jbody, "Ip": ctx.GetIp()}
			strlog, _ := json.Marshal(&jlog)
			this.token.RPush(fmt.Sprintf("%s:token:http_request", project), string(strlog))
		}
		handler(ctx)
	})
}

func (this *XHttp) SetToken(key string, data interface{}) {
	if this.token == nil {
		return
	}
	this.token.Set(fmt.Sprintf("%v:token:%s", project, key), data, 86400*7)
}

func (this *XHttp) DelToken(key string) {
	if this.token == nil {
		return
	}
	if key == "" {
		return
	}
	this.token.Del(fmt.Sprintf("%v:token:%s", project, key))
}

func (this *XHttp) GetToken(key string) interface{} {
	if this.token == nil {
		return nil
	}
	data, _ := this.token.Get(fmt.Sprintf("%v:token:%s", project, key))
	return data
}

func (this *XHttp) RenewToken(key string) {
	if this.token == nil {
		return
	}
	this.token.Expire(fmt.Sprintf("%v:token:%s", project, key), 86400*7)
}
