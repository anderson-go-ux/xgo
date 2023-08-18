package xgo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/beego/beego/logs"
	"github.com/google/uuid"
)

var thisdb *XDb
var thisredis *XRedis
var thishttp *XHttp

type StructConfig struct {
	ChannelId   int
	ConfigName  string
	ConfigValue string
}
type AdminModifyConfigData struct {
	SellerId int `validate:"required" `
	Config   []StructConfig
}

var beforeModifyConfig func(*AdminModifyConfigData)
var afterAddChannel func(int)

type AdminTokenData struct {
	Account   string
	SellerId  int
	ChannelId int
	AuthData  string
}

func AdminBeforeModifyConfig(cb func(*AdminModifyConfigData)) {
	beforeModifyConfig = cb
}

func AdminAfterAddChannel(cb func(int)) {
	afterAddChannel = cb
}

func AdminInit(http *XHttp, db *XDb, redis *XRedis, fullauth string) {
	thishttp = http
	thisdb = db
	thisredis = redis
	if env != "dev" {
		data, _ := db.Table("x_seller").Select("count(*) as count").GetOne()
		count := data.GetInt("count")
		if count == 0 {
			db.Table("x_seller").Insert(H{
				"SellerId":   1,
				"SellerName": "初始运营商",
			})
		}
		data, _ = db.Table("x_channel").Select("count(*) as count").GetOne()
		count = data.GetInt("count")
		if count == 0 {
			db.Table("x_channel").Insert(H{
				"SellerId":    1,
				"ChannelId":   1,
				"ChannelName": "初始渠道",
			})
		}
		auth_init(db, fullauth)
	}
	/*
		drop table x_seller;
		drop table x_channel;
		drop table x_admin_user;
		drop table x_admin_role;
		drop table x_admin_opt_log;
		drop table x_admin_login_log;
	*/
	http.OnPostNoAuth("/sapi/user_login", user_login)
	http.OnPostNoAuth("/sapi/user_logout", user_logout)
	http.OnPostWithAuth("/sapi/get_seller_names", get_seller_names, "", false, "")
	http.OnPostWithAuth("/sapi/get_channel_names", get_channel_names, "", false, "")
	http.OnPostWithAuth("/sapi/get_role_names", get_role_names, "", false, "")
	http.OnPostWithAuth("/sapi/get_role_data", get_role_data, "", false, "")

	http.OnPostWithAuth("/sapi/get_channel", get_channel, "系统管理.渠道管理.查", false, "")
	http.OnPostWithAuth("/sapi/add_channel", add_channel, "系统管理.渠道管理.增", true, "添加渠道")
	http.OnPostWithAuth("/sapi/modify_channel", modify_channel, "系统管理.渠道管理.改", true, "修改渠道")
	http.OnPostWithAuth("/sapi/delete_channel", delete_channel, "系统管理.渠道管理.删", true, "删除渠道")

	http.OnPostWithAuth("/sapi/get_role", get_role, "系统管理.角色管理.查", false, "")
	http.OnPostWithAuth("/sapi/add_role", add_role, "系统管理.角色管理.增", true, "添加角色")
	http.OnPostWithAuth("/sapi/modify_role", modify_role, "系统管理.角色管理.改", true, "修改角色")
	http.OnPostWithAuth("/sapi/delete_role", delete_role, "系统管理.角色管理.删", true, "删除角色")

	http.OnPostWithAuth("/sapi/get_admin_user", get_admin_user, "系统管理.账号管理.查", false, "")
	http.OnPostWithAuth("/sapi/add_admin_user", add_admin_user, "系统管理.账号管理.增", true, "添加账号")
	http.OnPostWithAuth("/sapi/modify_admin_user", modify_admin_user, "系统管理.账号管理.改", true, "修改账号")
	http.OnPostWithAuth("/sapi/delete_admin_user", delete_admin_user, "系统管理.账号管理.删", true, "删除账号")
	http.OnPostWithAuth("/sapi/modify_admin_user_google", modify_admin_user_google, "系统管理.账号管理.改", true, "修改管理员验证码")

	http.OnPostWithAuth("/sapi/get_login_log", get_login_log, "系统管理.登录日志.查", false, "")
	http.OnPostWithAuth("/sapi/get_opt_log", get_opt_log, "系统管理.操作日志.查", false, "")

	http.OnPostWithAuth("/sapi/get_system_config", get_system_config, "系统管理.系统设置.查", false, "")
	http.OnPostWithAuth("/sapi/modify_system_config", modify_system_config, "系统管理.系统设置.改", false, "")
}

func GetAdminToken(ctx *XHttpContent) *AdminTokenData {
	tokendata := AdminTokenData{}
	err := json.Unmarshal([]byte(ctx.TokenData), &tokendata)
	if err != nil {
		return nil
	}
	return &tokendata
}

func auth_init(db *XDb, fullauth string) {
	jdata := map[string]interface{}{}
	json.Unmarshal([]byte(fullauth), &jdata)
	xitong := jdata["系统管理"].(map[string]interface{})
	xitong["运营商管理"] = map[string]interface{}{"查:": 0, "增": 0, "删": 0, "改": 0}
	if xitong["系统设置"] != nil {
		xitongsezhi := xitong["系统设置"].(map[string]interface{})
		xitongsezhi["删"] = 0
	}
	jbytes, _ := json.Marshal(&jdata)
	authstr := string(jbytes)
	sellers, err := thisdb.Table("x_seller").GetList()
	sellers.ForEach(func(xd *XDbData) bool {
		SellerId := xd.GetInt("SellerId")
		_, err := thisdb.Table("x_admin_role").Where("SellerId = ? and RoleName = '运营商超管'", SellerId, nil).GetOne()
		if err != nil && err.Error() == DB_ERROR_NORECORD {
			thisdb.Table("x_admin_role").Insert(H{
				"SellerId": SellerId,
				"Parent":   "god",
				"RoleName": "运营商超管",
				"RoleData": authstr,
			})
		}
		_, err = thisdb.Table("x_admin_user").Where("SellerId = ?", SellerId, nil).GetOne()
		if err != nil && err.Error() == DB_ERROR_NORECORD {
			thisdb.Table("x_admin_user").Insert(H{
				"SellerId": SellerId,
				"Account":  fmt.Sprintf("admin%v", SellerId),
				"Password": Md5(Md5("admin")),
				"RoleName": "运营商超管",
			})

		}
		return true
	})
	sql := "update x_admin_role set RoleData = ? where RoleName = ?"
	db.conn().Exec(sql, authstr, "运营商超管")
	super, err := thisdb.Table("x_admin_role").Where("SellerId = ? and RoleName = '超级管理员'", -1, nil).GetOne()
	if super.GetString("RoleData") != fullauth {
		roles, _ := thisdb.Table("x_admin_role").GetList()
		roles.ForEach(func(xd *XDbData) bool {
			if xd.GetString("RoleName") == "超级管理员" {
				return true
			}
			jnewdata := make(map[string]interface{})
			json.Unmarshal([]byte(fullauth), &jnewdata)
			clean_auth(jnewdata)
			jrdata := make(map[string]interface{})
			json.Unmarshal([]byte(xd.GetString("RoleData")), &jrdata)
			for k, v := range jrdata {
				set_auth(k, jnewdata, v.(map[string]interface{}))
			}
			newauthbyte, _ := json.Marshal(&jnewdata)
			sql = "update x_admin_role set RoleData = ? where id = ?"
			thisdb.Exec(sql, string(newauthbyte), xd.GetInt("Id"))
			return true
		})
	}
	if err != nil && err.Error() == DB_ERROR_NORECORD {
		thisdb.Table("x_admin_role").Insert(H{
			"SellerId": -1,
			"RoleName": "超级管理员",
			"Parent":   "god",
			"RoleData": fullauth,
		})
	} else {
		sql = "update x_admin_role set RoleData = ? where RoleName = ?"
		thisdb.Exec(sql, fullauth, "超级管理员")
	}
}

func clean_auth(node map[string]interface{}) {
	for k, v := range node {
		if strings.Index(reflect.TypeOf(v).Name(), "float") >= 0 {
			node[k] = 0
		} else {
			clean_auth(v.(map[string]interface{}))
		}
	}
}

func set_auth(parent string, newdata map[string]interface{}, node map[string]interface{}) {
	for k, v := range node {
		if strings.Index(reflect.TypeOf(v).Name(), "float") >= 0 {
			if InterfaceToFloat(v) != 1 {
				continue
			}
			path := strings.Split(parent, ".")
			if len(path) == 0 {
				continue
			}
			fk, fok := newdata[path[0]]
			if !fok {
				continue
			}
			var pn *interface{} = &fk
			var finded bool = true
			for i := 1; i < len(path); i++ {
				tk := path[i]
				tv, ok := (*pn).(map[string]interface{})[tk]
				if !ok {
					finded = false
					break
				}
				pn = &tv
			}
			if finded {
				(*pn).(map[string]interface{})[k] = 1
			}
		} else {
			set_auth(parent+"."+k, newdata, v.(map[string]interface{}))
		}
	}
}

func user_login(ctx *XHttpContent) {
	type RequestData struct {
		Account    string `validate:"required"`
		Password   string `validate:"required"`
		GoogleCode string `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	reqdata.Password = Md5(reqdata.Password)
	type MenuData struct {
		Icon  string     `json:"icon"`
		Index string     `json:"index"`
		Title string     `json:"title"`
		Subs  []MenuData `json:"subs"`
	}
	lockkey := fmt.Sprintf("lock:admin_login:%v", reqdata.Account)
	if !thisredis.GetLock(lockkey, 10) {
		ctx.RespErr("操作频繁,请稍后再试")
		return
	}
	user, err := thisdb.Table("x_admin_user").Where("Account = ?", reqdata.Account, nil).GetOne()
	if err != nil {
		if err.Error() == "record not found" {
			ctx.RespErr("账号不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
		return
	}
	if user.GetInt("State") != 1 {
		ctx.RespErr("账号已禁用")
		return
	}
	if strings.Index(env, "prd") >= 0 && user.GetString("LoginGoogle") != "" && !VerifyGoogleCode(user.GetString("LoginGoogle"), reqdata.GoogleCode) {
		ctx.RespErr("验证码错误")
		return
	}
	if user.GetString("Password") != reqdata.Password {
		ctx.RespErr("密码不正确")
		return
	}

	seller, err := thisdb.Table("x_seller").Where("SellerId = ?", user.GetString("SellerId"), nil).GetOne()
	if err != nil {
		if err.Error() == DB_ERROR_NORECORD {
			ctx.RespErr("运营商不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
		return
	}
	if seller.GetInt("State") != 1 {
		ctx.RespErr("运营商已禁用")
		return
	}
	role, err := thisdb.Table("x_admin_role").Where("SellerId = ?", user.GetInt("SellerId"), nil).
		Where("RoleName = ?", user.GetString("RoleName"), nil).GetOne()
	if err != nil {
		if err.Error() == DB_ERROR_NORECORD {
			ctx.RespErr("角色不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
	}
	if role.GetInt("State") != 1 {
		ctx.RespErr("角色已禁用")
		return
	}

	if len(user.GetString("Token")) > 0 {
		thishttp.DelToken(user.GetString("Token"))
	}
	if user.GetInt("ChannelId") != 0 {
		ctx.RespErr("该账号为渠道账号,不可登录运营后台")
		return
	}
	token := uuid.New().String()
	tokendata := AdminTokenData{}
	tokendata.Account = reqdata.Account
	tokendata.SellerId = user.GetInt("SellerId")
	tokendata.ChannelId = user.GetInt("ChannelId")
	tokendata.AuthData = role.GetString("RoleData")
	thishttp.SetToken(token, tokendata)
	sql := "update x_admin_user set Token = ?,LoginCount = LoginCount + 1,LoginTime = now(),LoginIp = ? where id = ?"
	thisdb.Exec(sql, token, ctx.GetIp(), user.GetInt("Id"))
	thisdb.Table("x_admin_login_log").Insert(H{
		"SellerId":   tokendata.SellerId,
		"ChannelId":  tokendata.ChannelId,
		"Account":    reqdata.Account,
		"Token":      token,
		"LoginIp":    ctx.GetIp(),
		"CreateTime": GetLocalTime(),
	})
	jauth := make(map[string]interface{})
	json.Unmarshal([]byte(tokendata.AuthData), &jauth)
	ctx.Put("UserId", user.GetInt("Id"))
	ctx.Put("SellerId", user.GetInt("SellerId"))
	ctx.Put("ChannelId", user.GetInt("ChannelId"))
	ctx.Put("Account", reqdata.Account)
	ctx.Put("Token", token)
	ctx.Put("LoginTime", user.GetString("LoginTime"))
	ctx.Put("Ip", ctx.GetIp())
	ctx.Put("LoginCount", user.GetInt("LoginCount"))
	ctx.Put("AuthData", jauth)
	ctx.RespOK()
}

func user_logout(ctx *XHttpContent) {
	type RequestData struct {
		Token string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.Token != "" {
		thishttp.DelToken(reqdata.Token)
	}
	ctx.RespOK()
}

func get_seller_names(ctx *XHttpContent) {
	sellers, _ := thisdb.Table("x_seller").Select("SellerId,SellerName").GetList()
	ctx.Put("data", sellers.GetData())
	ctx.RespOK()
}

func get_channel_names(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_channel")
	table = table.Select("ChannelId,ChannelName")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.OrderBy("ChannelId asc")
	channels, _ := table.GetList()
	ctx.Put("data", channels.GetData())
	ctx.RespOK()
}

func get_role_names(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_role")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Select("RoleName")
	roles, _ := table.GetList()
	ctx.Put("data", roles.GetData())
	ctx.RespOK()
}

func get_role_data(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		RoleName string `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	{
		table := thisdb.Table("x_admin_role")
		table = table.Where("SellerId = ?", reqdata.SellerId, nil)
		table = table.Where("RoleName = ?", reqdata.RoleName, nil)
		table = table.Select("RoleData")
		role, _ := table.GetOne()
		if role != nil {
			ctx.Put("RoleData", role.GetString("RoleData"))
		}
	}
	{
		table := thisdb.Table("x_admin_role")
		table = table.Where("SellerId = ?", reqdata.SellerId, nil)
		table = table.Where("RoleName = ?", "运营商超管", nil)
		table = table.Select("RoleData")
		role, _ := table.GetOne()
		if role != nil {
			ctx.Put("SuperRoleData", role.GetString("RoleData"))
		}
	}
	ctx.RespOK()
}

func get_channel(ctx *XHttpContent) {
	type RequestData struct {
		Page        int    `gorm:"-"`
		PageSize    int    `gorm:"-"`
		SellerId    int    `gorm:"column:SellerId"`
		ChannelId   int    `gorm:"column:ChannelId"`
		ChannelName string `gorm:"column:ChannelName"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.Page <= 0 {
		reqdata.Page = 1
	}
	if reqdata.PageSize <= 0 || reqdata.PageSize > 2000 {
		reqdata.PageSize = 15
	}
	table := thisdb.Table("x_channel").OrderBy("id desc")
	table = table.Where("SellerId = ?", reqdata.SellerId, 0)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, 0)
	table = table.Where("ChannelName = ?", reqdata.ChannelName, "")
	total, _ := table.Count("")
	channels, _ := table.GetList()
	ctx.Put("data", channels.GetData())
	ctx.Put("total", total)
	ctx.RespOK()
}

func add_channel(ctx *XHttpContent) {
	type RequestData struct {
		SellerId    int    `validate:"required" gorm:"column:SellerId"`
		ChannelId   int    `validate:"required" gorm:"column:ChannelId"`
		ChannelName string `validate:"required" gorm:"column:ChannelName"`
		Memo        string `gorm:"column:Memo"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_channel")
	_, err := table.Insert(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("添加失败")
		return
	}
	if afterAddChannel != nil {
		afterAddChannel(reqdata.ChannelId)
	}
	ctx.RespOK()
}

func modify_channel(ctx *XHttpContent) {
	type RequestData struct {
		SellerId    int `validate:"required"`
		ChannelId   int `validate:"required"`
		ChannelName string
		State       int
		Memo        string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_channel")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, nil)
	_, err := table.Update(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("修改失败")
		return
	}
	ctx.RespOK()
}

func delete_channel(ctx *XHttpContent) {
	type RequestData struct {
		SellerId  int `validate:"required" gorm:"column:SellerId"`
		ChannelId int `validate:"required" gorm:"column:ChannelId"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_channel")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, nil)
	_, err := table.Delete()
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("删除失败")
		return
	}
	ctx.RespOK()
}

func get_role(ctx *XHttpContent) {
	type RequestData struct {
		Page     int
		PageSize int
		SellerId int `validate:"required"`
		RoleName string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_role").OrderBy("id desc")
	table = table.Where("SellerId = ?", reqdata.SellerId, 0)
	table = table.Where("RoleName = ?", reqdata.RoleName, "")
	total, _ := table.Count("")
	roles, _ := table.PageData(reqdata.Page, reqdata.PageSize)
	ctx.Put("data", roles.GetData())
	ctx.Put("total", total)
	ctx.RespOK()
}

func add_role(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		RoleName string `validate:"required"`
		Parent   string `validate:"required"`
		RoleData string `validate:"required"`
		Memo     string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	db := thisdb.Table("x_admin_role")
	_, err := db.Insert(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("添加失败")
		return
	}
	ctx.RespOK()
}

func modify_role(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		RoleName string `validate:"required"`
		Parent   string `validate:"required"`
		RoleData string
		State    int
		Memo     string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_role")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("RoleName = ?", reqdata.RoleName, nil)
	_, err := table.Update(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("修改失败")
		return
	}
	ctx.RespOK()
}

func delete_role(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		RoleName string `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_role")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("RoleName = ?", reqdata.RoleName, nil)
	_, err := table.Delete()
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("删除失败")
		return
	}
	ctx.RespOK()
}

func get_admin_user(ctx *XHttpContent) {
	type RequestData struct {
		Page      int
		PageSize  int
		SellerId  int `validate:"required"`
		ChannelId int
		Account   string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.Page <= 0 {
		reqdata.Page = 1
	}
	if reqdata.PageSize <= 0 || reqdata.PageSize > 2000 {
		reqdata.PageSize = 15
	}
	table := thisdb.Table("x_admin_User")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, 0)
	table = table.Where("Account = ?", reqdata.Account, "")
	total, _ := table.Select("select count(*) as total").GetOne()
	users, _ := table.Select("*").PageData(reqdata.Page, reqdata.PageSize)
	users.ForEach(func(xd *XDbData) bool {
		xd.Delete("Token")
		xd.Delete("Password")
		xd.Delete("LoginGoogle")
		xd.Delete("OptGoogle")
		return true
	})
	ctx.Put("data", users.GetData())
	ctx.Put("total", total.GetInt64("total"))
	ctx.RespOK()
}

func add_admin_user(ctx *XHttpContent) {
	type RequestData struct {
		SellerId  int `validate:"required"`
		ChannelId int
		Account   string `validate:"required"`
		Password  string `validate:"required"`
		RoleName  string `validate:"required"`
		Memo      string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	reqdata.Password = Md5(reqdata.Password)
	db := thisdb.Table("x_admin_user")
	_, err := db.Insert(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("添加失败")
		return
	}
	ctx.RespOK()
}

func modify_admin_user(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		Account  string `validate:"required"`
		Password string
		State    int
		Memo     string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.Password != "" {
		reqdata.Password = Md5(reqdata.Password)
	}
	table := thisdb.Table("x_admin_user")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("Account = ?", reqdata.Account, nil)
	_, err := table.Update(reqdata)
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("修改失败")
		return
	}
	ctx.RespOK()
}

func delete_admin_user(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required"`
		Account  string `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_user")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("Account = ?", reqdata.Account, nil)
	_, err := table.Delete()
	if err != nil {
		logs.Error(err)
		ctx.Put("error", err.Error())
		ctx.RespErr("删除失败")
		return
	}
	ctx.RespOK()
}

func modify_admin_user_google(ctx *XHttpContent) {
	type RequestData struct {
		SellerId   int    `validate:"required" `
		Account    string `validate:"required"`
		CodeType   int    `validate:"required" `
		GoogleCode string `validate:"required" `
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	token := GetAdminToken(ctx)

	lockkey := fmt.Sprintf("lock:admin_change_google:%v", reqdata.Account)
	if !thisredis.GetLock(lockkey, 10) {
		ctx.RespErr("操作频繁,请稍后再试")
		return
	}

	user, err := thisdb.Table("x_admin_user").Where("SellerId = ?", reqdata.SellerId, nil).Where("Account = ?", reqdata.Account, nil).GetOne()
	if err != nil && err.Error() == DB_ERROR_NORECORD {
		ctx.RespErr("管理员不存在")
		return
	}

	me, _ := thisdb.Table("x_admin_user").Where("SellerId = ?", token.SellerId, nil).Where("Account = ?", token.Account, nil).GetOne()
	OptGoogle := me.GetString("OptGoogle")
	LoginGoogle := me.GetString("LoginGoogle")
	if reqdata.Account != token.Account {
		if OptGoogle == "" {
			ctx.RespErr(fmt.Sprintf("请先设置账号 %v 的操作验证码", token.Account))
			return
		}
		if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(OptGoogle, reqdata.GoogleCode) {
			ctx.RespErr("验证码不正确")
			return
		}
	} else {
		if reqdata.CodeType == 2 && LoginGoogle == "" {
			if OptGoogle == "" {
				ctx.RespErr(fmt.Sprintf("请先设置账号 %v 的登录验证码", me.GetString("Account")))
				return
			}
		}
		if OptGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(OptGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的操作验证码", me.GetString("Account")))
				return
			}
		}
		if LoginGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(LoginGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的登录 验证码", me.GetString("Account")))
				return
			}
		}
	}
	seller, _ := thisdb.Table("x_seller").Where("SellerId = ?", reqdata.SellerId, nil).GetOne()
	if reqdata.CodeType == 1 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-登录", reqdata.Account, verifykey, seller.GetString("SellerName"))
		thisdb.Exec("update x_admin_user set LoginGoogle = ? where Id = ?", verifykey, user.GetInt("Id"))
		ctx.Put("url", verifyurl)
		ctx.RespOK()
	} else if reqdata.CodeType == 2 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-操作", reqdata.Account, verifykey, seller.GetString("SellerName"))
		thisdb.Exec("update x_admin_user set OptGoogle = ? where Id = ?", verifykey, user.GetInt("Id"))
		ctx.Put("url", verifyurl)
		ctx.RespOK()
	}
}

func get_login_log(ctx *XHttpContent) {
	type RequestData struct {
		Page      int
		PageSize  int
		SellerId  int `validate:"required" `
		ChannelId int
		Account   string
		LoginIp   string
		StartTime string
		EndTime   string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.StartTime != "" {
		reqdata.StartTime = UtcToLocalTime(reqdata.StartTime)
	}
	if reqdata.EndTime != "" {
		reqdata.EndTime = UtcToLocalTime(reqdata.EndTime)
	}
	table := thisdb.Table("x_admin_login_log")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, 0)
	table = table.Where("Account = ?", reqdata.Account, "")
	table = table.Where("LoginIp = ?", reqdata.LoginIp, "")
	table = table.Where("CreateTime >= ?", reqdata.StartTime, "")
	table = table.Where("CreateTime < ?", reqdata.EndTime, "")
	total, _ := table.Select("count(*) as total").GetOne()
	logs, _ := table.Select("*").PageData(reqdata.Page, reqdata.PageSize)
	ctx.Put("data", logs.GetData())
	ctx.Put("total", total.GetInt64("total"))
	ctx.RespOK()
}

func get_opt_log(ctx *XHttpContent) {
}

func get_system_config(ctx *XHttpContent) {
	type RequestData struct {
		SellerId   int `validate:"required" `
		ChannelId  int
		ConfigName []string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_config")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	if reqdata.ChannelId == 0 {
		table = table.Where("ChannelId = ?", 0, nil)
	} else {
		table = table.Where("(ChannelId = 0 or ChannelId = ?)", reqdata.ChannelId, nil)
	}
	if len(reqdata.ConfigName) > 0 {
		table = table.Where("ConfigName in  ", reqdata.ConfigName, nil)
	}
	config, _ := table.GetList()
	ctx.Put("data", config.GetData())
	ctx.RespOK()
}

func modify_system_config(ctx *XHttpContent) {
	reqdata := AdminModifyConfigData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if beforeModifyConfig != nil {
		beforeModifyConfig(&reqdata)
	}
	for i := 0; i < len(reqdata.Config); i++ {
		sql := "update x_config set ConfigValue = ? where SellerId = ? and ChannelId = ? and ConfigName = ?"
		thisdb.Exec(sql, reqdata.Config[i].ConfigValue, reqdata.SellerId, reqdata.Config[i].ChannelId, reqdata.Config[i].ConfigName)
	}
	ctx.RespOK()
}
