package xgo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/beego/beego/logs"
	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

type RequestLogCallback func([]byte)

type XHttp struct {
	gin   *gin.Engine
	token *XRedis

	upgrader             websocket.Upgrader
	request_log_callback RequestLogCallback
	idx_conn             sync.Map
	conn_idx             sync.Map
	connect_callback     XWsCallback
	close_callback       XWsCallback
	msgtype              sync.Map
	msg_callback         sync.Map
	default_msg_callback XWsDefaultMsgCallback
	logwriter            io.Writer
	logname              string
	port                 int
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

// 获取请求参数
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

// 获取url请求参数
func (this *XHttpContent) Query(key string) string {
	return this.gin.Query(key)
}

// 获取客户端ip
func (this *XHttpContent) GetIp() string {
	return this.gin.ClientIP()
}

// 返回gin上下文
func (this *XHttpContent) Gin() *gin.Context {
	return this.gin
}

// 获取请求域名
func (c *XHttpContent) Host() string {
	return strings.Split(c.gin.Request.Host, ":")[0]
}

// 设置返回数据
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

// 返回成功
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

// 返回json
func (ctx *XHttpContent) RespJson(obj any) {
	ctx.gin.JSON(http.StatusOK, obj)
}

// 返回失败
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

// 初始化网络服务
func (this *XHttp) Init(cfgname string, token *XRedis) {
	this.token = token
	this.port = int(GetConfigInt(cfgname+".port", true, 0))
	this.gin = gin.New()
	this.gin.Use(abuhttpcors())
	go func() {
		bind := fmt.Sprint("0.0.0.0:", this.port)
		this.gin.Run(bind)
	}()
	this.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	if this.token != nil && this.request_log_callback != nil {
		go func() {
			for {
				logdata, _ := this.token.BLPop("token:http_request", 86400)
				if logdata != nil && this.request_log_callback != nil {
					this.request_log_callback(logdata)
				}
			}
		}()
	}
	logs.Debug("http listen:", this.port)
}

// 获取监听端口
func (this *XHttp) Port() int {
	return this.port
}

// 设置日志回调
func (this *XHttp) SetLogCallback(cb RequestLogCallback) {
	this.request_log_callback = cb
}

// 设置静态访问路径
func (this *XHttp) Static(relativePaths string, root string) {
	this.gin.Static(relativePaths, root)
}

// 响应post请求
func (this *XHttp) OnPost(path string, handler XHttpHandler) {
	this.OnPostWithAuth(path, handler, "", false, "")
}

// 响应post请求,同时验证权限,谷歌验证码
func (this *XHttp) OnPostWithAuth(path string, handler XHttpHandler, auth string, googleverify bool, optname string) {
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
		rediskey := fmt.Sprintf("token:%v", tokenstr)
		tokendata, _ := this.token.Get(rediskey)
		if tokendata == nil {
			ctx.RespErr(3, "未登录或登录已过期")
			return
		}
		keystr := fmt.Sprintf("post%v%v", gc.Request.URL.Path, tokenstr)
		reqid := Md5(keystr)
		lockkey := fmt.Sprintf("token:lock:%v", reqid)
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
				ctx.RespErr(8, "请填写验证码")
				return
			}
			gcstr := ToString(gc)
			if len(gcstr) == 0 {
				ctx.RespErr(8, "请填写验证码")
				return
			}
			if strings.Index(env, "prd") > 0 {
				gsstr := ToString(jtoken["GoogleSecret"])
				if !VerifyGoogleCode(gsstr, gcstr) {
					ctx.RespErr(9, "验证码不正确")
					return
				}
			}
		}
		jlog := H{
			"ReqPath":   gc.Request.URL.Path,
			"ReqData":   jbody,
			"Account":   jtoken["Account"],
			"UserId":    jtoken["UserId"],
			"SellerId":  jtoken["SellerId"],
			"ChannelId": jtoken["ChannelId"],
			"Ip":        ctx.GetIp(),
			"Token":     tokenstr,
			"OptName":   optname,
			"OptTime":   GetLocalTime(),
		}
		strlog, _ := json.Marshal(&jlog)
		this.token.RPush("token:http_request", string(strlog))
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
				if ToInt(io) != 1 {
					ctx.RespErr(5, "权限不足")
					return
				}
			}
		}
		go func() {
			defer func() {
				recover()
			}()
			filename := fmt.Sprintf("_log/gin_%v.log", GetLocalDate())
			if this.logname != filename {
				this.logname = filename
				file, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, os.ModePerm)
				this.logwriter = io.MultiWriter(file)
			}
			fbody, _ := json.Marshal(jbody)
			text := fmt.Sprintf("[%v][%v][%v][%v]\r\n", GetLocalTime(), ctx.GetIp(), gc.Request.URL.Path, string(fbody))
			this.logwriter.Write([]byte(text))
		}()
		handler(ctx)
	})
}

// 响应post请求,无需权限,谷歌验证
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
		jbody := map[string]interface{}{}
		err := json.Unmarshal([]byte(strbody), &jbody)
		if err != nil {
			ctx.RespErr(4, "参数必须是json格式")
			return
		}
		jlog := gin.H{"ReqPath": gc.Request.URL.Path, "ReqData": jbody, "Ip": ctx.GetIp(), "OptTime": GetLocalTime()}
		strlog, _ := json.Marshal(&jlog)
		if this.token != nil {
			this.token.RPush("token:http_request", string(strlog))
		}
		go func() {
			filename := fmt.Sprintf("_log/gin_%v.log", GetLocalDate())
			if this.logname != filename {
				this.logname = filename
				file, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, os.ModePerm)
				this.logwriter = io.MultiWriter(file)
			}
			fbody, _ := json.Marshal(jbody)
			text := fmt.Sprintf("[%v][%v][%v][%v]\r\n", GetLocalTime(), ctx.GetIp(), gc.Request.URL.Path, string(fbody))
			this.logwriter.Write([]byte(text))
		}()
		handler(ctx)
	})
}

// 保存token
func (this *XHttp) SetToken(key string, data interface{}) {
	if this.token == nil {
		return
	}
	this.token.Set(fmt.Sprintf("token:%s", key), data, 86400*7)
}

// 删除token
func (this *XHttp) DelToken(key string) {
	if this.token == nil {
		return
	}
	if key == "" {
		return
	}
	this.token.Del(fmt.Sprintf("token:%s", key))
}

// 获取token数据
func (this *XHttp) GetToken(key string) interface{} {
	if this.token == nil {
		return nil
	}
	data, _ := this.token.Get(fmt.Sprintf("token:%s", key))
	return data
}

// token延时续约
func (this *XHttp) RenewToken(key string) {
	if this.token == nil {
		return
	}
	this.token.Expire(fmt.Sprintf("token:%s", key), 86400*7)
}

type XWsCallback func(string)
type XWsMsgCallback func(string, interface{})
type XWsDefaultMsgCallback func(string, string, interface{})

type abumsgdata struct {
	MsgId string      `json:"msgid"`
	Data  interface{} `json:"data"`
}

// 初始化ws
func (this *XHttp) InitWs(url string) {
	this.gin.GET(url, func(gc *gin.Context) {
		ctx := &XHttpContent{gc, "", "", ""}
		this.ws(ctx)
	})
}

func (this *XHttp) ws(ctx *XHttpContent) {
	conn, err := this.upgrader.Upgrade(ctx.Gin().Writer, ctx.Gin().Request, nil)
	if err != nil {
		logs.Error(err)
		return
	}
	defer conn.Close()
	id := uuid.New().String()
	id = strings.Replace(id, "-", "", -1)
	this.idx_conn.Store(id, conn)
	this.conn_idx.Store(conn, id)
	if this.connect_callback != nil {
		this.connect_callback(id)
	}
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		this.msgtype.Store(id, mt)
		md := abumsgdata{}
		err = json.Unmarshal(message, &md)
		if err != nil {
			break
		}
		if len(md.MsgId) == 0 {
			break
		}
		callback, cbok := this.msg_callback.Load(md.MsgId)
		if cbok {
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						logs.Error(err)
						stack := debug.Stack()
						logs.Error(string(stack))
					}
				}()
				cb := callback.(XWsMsgCallback)
				cb(id, md.Data)
			}()
		} else {
			if this.default_msg_callback != nil {
				go func() {
					defer func() {
						err := recover()
						if err != nil {
							logs.Error(err)
							stack := debug.Stack()
							logs.Error(string(stack))
						}
					}()
					this.default_msg_callback(id, md.MsgId, md.Data)
				}()
			}
		}
	}
	this.idx_conn.Delete(id)
	this.conn_idx.Delete(conn)
	mt, merr := this.msgtype.Load(id)
	if merr {
		this.msg_callback.Delete(mt)
	}
	_, ccerr := this.idx_conn.Load(id)
	if ccerr && this.close_callback != nil {
		this.close_callback(id)
	}
}

// ws连接回调
func (this *XHttp) OnWsConnect(cb XWsCallback) {
	this.connect_callback = cb
}

// ws关闭回调
func (this *XHttp) OnWsClose(cb XWsCallback) {
	this.close_callback = cb
}

// ws消息回调
func (this *XHttp) OnWsMessage(msgid string, cb XWsMsgCallback) {
	this.msg_callback.Store(msgid, cb)
}

// 移除ws消息回调
func (this *XHttp) RemoveWsMessage(msgid string) {
	this.msg_callback.Delete(msgid)
}

// ws默认消息回调
func (this *XHttp) OnWsDefaultMessage(cb XWsDefaultMsgCallback) {
	this.default_msg_callback = cb
}

// 关闭ws
func (this *XHttp) CloseWs(id string) {
	conn, ok := this.idx_conn.Load(id)
	if ok {
		conn.(*websocket.Conn).Close()
	}
}

// 发送ws消息
func (this *XHttp) SendWsMsg(id string, msgid string, data interface{}) {
	conn, ok := this.idx_conn.Load(id)
	if ok {
		msg := abumsgdata{MsgId: msgid, Data: data}
		msgdata, _ := json.Marshal(&msg)
		mt, mok := this.msgtype.Load(id)
		if !mok {
			mt = websocket.TextMessage
		}
		conn.(*websocket.Conn).WriteMessage(mt.(int), msgdata)
	}
}

// 发送ws消息
func (this *XHttp) SendWsData(id string, data interface{}) {
	conn, ok := this.idx_conn.Load(id)
	if ok {
		mt, mok := this.msgtype.Load(id)
		if !mok {
			mt = websocket.TextMessage
		}
		conn.(*websocket.Conn).WriteMessage(mt.(int), []byte(ToString(data)))
	}
}
