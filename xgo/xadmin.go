package xgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/beego/beego/logs"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var thisdb *XDb
var thisredis *XRedis
var thishttp *XHttp

type StructConfig struct {
	ChannelId   int
	ConfigName  string
	ConfigValue string
}
type ModifyConfigData struct {
	SellerId int `validate:"required" `
	Config   []StructConfig
}

var beforeModifyConfig func(ModifyConfigData) ModifyConfigData
var afterAddChannel func(int)

type AdminTokenData struct {
	Account   string
	SellerId  int
	ChannelId int
	AuthData  string
}

type XAdminRole struct {
	Id         int    `gorm:"column:Id"`
	SellerId   int    `gorm:"column:SellerId"`
	RoleName   string `gorm:"column:RoleName"`
	Parent     string `gorm:"column:Parent"`
	RoleData   string `gorm:"column:RoleData"`
	State      int    `gorm:"column:State"`
	Memo       string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type XAdminUser struct {
	Id          int    `gorm:"column:Id"`
	SellerId    int    `gorm:"column:SellerId"`
	ChannelId   int    `gorm:"column:ChannelId"`
	Account     string `gorm:"column:Account"`
	Password    string `gorm:"column:Password"`
	RoleName    string `gorm:"column:RoleName"`
	LoginGoogle string `gorm:"column:LoginGoogle"`
	OptGoogle   string `gorm:"column:OptGoogle"`
	State       int    `gorm:"column:State"`
	Token       string `gorm:"column:Token"`
	LoginCount  int    `gorm:"column:LoginCount"`
	LoginTime   string `gorm:"column:LoginTime"`
	LoginIp     string `gorm:"column:LoginIp"`
	Memo        string `gorm:"column:Memo"`
	CreateTime  string `gorm:"column:CreateTime"`
}

type XAdminLoginLog struct {
	Id         int    `gorm:"column:Id"`
	SellerId   int    `gorm:"column:SellerId"`
	ChannelId  int    `gorm:"column:ChannelId"`
	Account    string `gorm:"column:Account"`
	Token      string `gorm:"column:Token"`
	LoginIp    string `gorm:"column:LoginIp"`
	Memo       string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type XAdminOptLog struct {
	Id         int    `gorm:"column:Id"`
	SellerId   int    `gorm:"column:SellerId"`
	ChannelId  int    `gorm:"column:ChannelId"`
	Account    string `gorm:"column:Account"`
	ReqPath    string `gorm:"column:ReqPath"`
	ReqData    string `gorm:"column:ReqData"`
	Ip         string `gorm:"column:Ip"`
	Memo       string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type XSeller struct {
	Id         int    `gorm:"column:Id"`
	SellerId   int    `gorm:"column:SellerId"`
	State      int    `gorm:"column:State"`
	SellerName string `gorm:"column:SellerName"`
	Memo       string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type XChannel struct {
	Id          int    `gorm:"column:Id"`
	SellerId    int    `gorm:"column:SellerId"`
	ChannelId   int    `gorm:"column:ChannelId"`
	State       int    `gorm:"column:State"`
	ChannelName string `gorm:"column:ChannelName"`
	Memo        string `gorm:"column:Memo"`
	CreateTime  string `gorm:"column:CreateTime"`
}

type XConfig struct {
	Id          int    `gorm:"column:Id"`
	SellerId    int    `gorm:"column:SellerId"`
	ChannelId   int    `gorm:"column:ChannelId"`
	ConfigName  string `gorm:"column:ConfigName"`
	ConfigValue string `gorm:"column:ConfigValue"`
	ForClient   int    `gorm:"column:ForClient"`
	CreateTime  string `gorm:"column:CreateTime"`
}

func AdminBeforeModifyConfig(cb func(ModifyConfigData) ModifyConfigData) {
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
		var count int
		db.Gorm().Table("x_seller").Count(&count)
		if count == 0 {
			seller := XSeller{}
			seller.SellerId = 1
			seller.SellerName = "初始运营商"
			db.Gorm().Table("x_seller").Create(&seller)
		}
		db.Gorm().Table("x_channel").Count(&count)
		if count == 0 {
			channel := XChannel{}
			channel.SellerId = 1
			channel.ChannelId = 1
			channel.ChannelName = "初始渠道"
			db.Gorm().Table("x_channel").Create(&channel)
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

	// http.OnPostWithAuth("/sapi/get_seller", get_seller, "系统管理.运营商管理.查", false, "")
	// http.OnPostWithAuth("/sapi/add_seller", add_seller, "系统管理.运营商管理.增", true, "添加运营商")
	// http.OnPostWithAuth("/sapi/modify_seller", modify_seller, "系统管理.运营商管理.改", true, "修改运营商")
	// http.OnPostWithAuth("/sapi/delete_seller", delete_seller, "系统管理.运营商管理.删", true, "删除运营商")

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
	sellers := []XSeller{}
	thisdb.Gorm().Table("x_seller").Find(&sellers)
	for i := 0; i < len(sellers); i++ {
		role := XAdminRole{}
		err := thisdb.Gorm().Table("x_admin_role").Where("SellerId = ? and RoleName = '运营商超管'", sellers[i].SellerId).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role.SellerId = sellers[i].SellerId
			role.Parent = "god"
			role.RoleName = "运营商超管"
			role.RoleData = authstr
			thisdb.Gorm().Table("x_admin_role").Create(&role)
		}
		user := XAdminUser{}
		err = thisdb.Gorm().Table("x_admin_user").Where("SellerId = ?", sellers[i].SellerId).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.SellerId = sellers[i].SellerId
			user.Account = fmt.Sprintf("admin%v", user.SellerId)
			user.Password = Md5(Md5("admin"))
			user.RoleName = "运营商超管"
			thisdb.Gorm().Table("x_admin_user").Create(&user)

		}
	}
	sql := "update x_admin_role set RoleData = ? where RoleName = ?"
	db.conn().Exec(sql, authstr, "运营商超管")

	super := XAdminRole{}
	err := thisdb.Gorm().Table("x_admin_role").Where("SellerId = -1 and RoleName = '超级管理员'").First(&super).Error
	if super.RoleData != fullauth {
		roles := []XAdminRole{}
		thisdb.Gorm().Table("x_admin_role").Find(&roles)
		for i := 0; i < len(roles); i++ {
			if roles[i].RoleName == "超级管理员" {
				continue
			}
			jnewdata := make(map[string]interface{})
			json.Unmarshal([]byte(fullauth), &jnewdata)
			clean_auth(jnewdata)
			jrdata := make(map[string]interface{})
			json.Unmarshal([]byte(roles[i].RoleData), &jrdata)
			for k, v := range jrdata {
				set_auth(k, jnewdata, v.(map[string]interface{}))
			}
			newauthbyte, _ := json.Marshal(&jnewdata)
			sql = "update x_admin_role set RoleData = ? where id = ?"
			thisdb.Exec(sql, string(newauthbyte), roles[i].Id)
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		super.SellerId = -1
		super.RoleName = "超级管理员"
		super.Parent = "god"
		super.RoleData = fullauth
		thisdb.Gorm().Table("x_admin_role").Create(&super)
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
	user := XAdminUser{}
	err := thisdb.Gorm().Table("x_admin_user").Where("Account = ?", reqdata.Account).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			ctx.RespErr("账号不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
		return
	}
	if user.State != 1 {
		ctx.RespErr("账号已禁用")
		return
	}
	if strings.Index(env, "prd") >= 0 && user.LoginGoogle != "" && !VerifyGoogleCode(user.LoginGoogle, reqdata.GoogleCode) {
		ctx.RespErr("验证码错误")
		return
	}
	if user.Password != reqdata.Password {
		ctx.RespErr("密码不正确")
		return
	}
	seller := XSeller{}
	err = thisdb.Gorm().Table("x_seller").Where("SellerId = ?", user.SellerId).First(&seller).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.RespErr("运营商不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
		return
	}
	if seller.State != 1 {
		ctx.RespErr("运营商已禁用")
		return
	}
	role := XAdminRole{}
	err = thisdb.Gorm().Table("x_admin_role").Where("SellerId = ? and RoleName = ?", user.SellerId, user.RoleName).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.RespErr("角色不存在")
		} else {
			logs.Error("user_login:", err)
			ctx.RespErr("登录失败,请稍后再试")
		}
	}
	if role.State != 1 {
		ctx.RespErr("角色已禁用")
		return
	}

	if len(user.Token) > 0 {
		thishttp.DelToken(user.Token)
	}
	if user.ChannelId != 0 {
		ctx.RespErr("该账号为渠道账号,不可登录运营后台")
		return
	}
	token := uuid.New().String()
	tokendata := AdminTokenData{}
	tokendata.Account = reqdata.Account
	tokendata.SellerId = user.SellerId
	tokendata.ChannelId = user.ChannelId
	tokendata.AuthData = role.RoleData
	thishttp.SetToken(token, tokendata)
	sql := "update x_admin_user set Token = ?,LoginCount = LoginCount + 1,LoginTime = now(),LoginIp = ? where id = ?"
	thisdb.Exec(sql, token, ctx.GetIp(), user.Id)
	log := XAdminLoginLog{}
	log.SellerId = user.SellerId
	log.ChannelId = user.ChannelId
	log.Account = reqdata.Account
	log.Token = token
	log.LoginIp = ctx.GetIp()
	log.CreateTime = GetLocalTime()
	thisdb.Gorm().Table("x_admin_login_log").Create(&log)
	jauth := make(map[string]interface{})
	json.Unmarshal([]byte(tokendata.AuthData), &jauth)
	ctx.Put("UserId", user.Id)
	ctx.Put("SellerId", user.SellerId)
	ctx.Put("ChannelId", user.ChannelId)
	ctx.Put("Account", reqdata.Account)
	ctx.Put("Token", token)
	ctx.Put("LoginTime", user.LoginTime)
	ctx.Put("Ip", ctx.GetIp())
	ctx.Put("LoginCount", user.LoginCount)
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
	seller := []XSeller{}
	db := thisdb.Gorm()
	db = db.Table("x_seller")
	db = db.Select("SellerId,SellerName")
	db.Find(&seller)
	sellers := []map[string]interface{}{}
	for i := 0; i < len(seller); i++ {
		sellers = append(sellers, H{"SellerId": seller[i].SellerId, "SellerName": seller[i].SellerName})
	}
	ctx.Put("data", sellers)
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
	channel := []XChannel{}
	db := thisdb.Gorm()
	db = db.Table("x_channel")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Select("ChannelId,ChannelName")
	db.Find(&channel)
	channels := []map[string]interface{}{}
	for i := 0; i < len(channel); i++ {
		channels = append(channels, H{"ChannelId": channel[i].ChannelId, "ChannelName": channel[i].ChannelName})
	}
	ctx.Put("data", channels)
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
	role := []XAdminRole{}
	db := thisdb.Gorm()
	db = db.Table("x_admin_role")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Select("RoleName")
	db.Find(&role)
	roles := []string{}
	for i := 0; i < len(role); i++ {
		roles = append(roles, role[i].RoleName)
	}
	ctx.Put("data", role)
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
		role := XAdminRole{}
		db := thisdb.Gorm()
		db = db.Table("x_admin_role")
		db = db.Where("SellerId = ?", reqdata.SellerId)
		db = db.Where("RoleName = ?", reqdata.RoleName)
		db = db.Select("RoleData")
		db.First(&role)
		ctx.Put("RoleData", role.RoleData)
	}
	{
		role := XAdminRole{}
		db := thisdb.Gorm()
		db = db.Table("x_admin_role")
		db = db.Where("SellerId = ?", reqdata.SellerId)
		db = db.Where("RoleName = ?", "运营商超管")
		db = db.Select("RoleData")
		db.First(&role)
		ctx.Put("SuperRoleData", role.RoleData)
	}
	ctx.RespOK()
}

// func get_seller(ctx *XHttpContent) {
// 	type RequestData struct {
// 		Page       int
// 		PageSize   int
// 		SellerId   int
// 		SellerName string
// 	}
// 	reqdata := RequestData{}
// 	if ctx.RequestData(&reqdata) != nil {
// 		return
// 	}
// 	if reqdata.Page <= 0 {
// 		reqdata.Page = 1
// 	}
// 	if reqdata.PageSize <= 0 || reqdata.PageSize > 2000 {
// 		reqdata.PageSize = 15
// 	}
// 	seller := []XSeller{}
// 	offset := (reqdata.Page - 1) * reqdata.PageSize
// 	db := thisdb.Gorm()
// 	db = thisdb.Where(db, "SellerId = ?", reqdata.SellerId, int(0))
// 	db = thisdb.Where(db, "SellerName = ?", reqdata.SellerName, "")
// 	db.Offset(offset).Limit(reqdata.PageSize).Find(&seller)
// 	for i := 0; i < len(seller); i++ {
// 		seller[i].CreateTime = LocalTimeToUtc(seller[i].CreateTime)
// 	}
// 	ctx.RespOK(seller)
// }

// func add_seller(ctx *XHttpContent) {
// 	type RequestData struct {
// 		SellerId   int    `validate:"required" gorm:"column:SellerId"`
// 		SellerName string `validate:"required" gorm:"column:SellerName"`
// 	}
// 	reqdata := RequestData{}
// 	if ctx.RequestData(&reqdata) != nil {
// 		return
// 	}
// 	db := thisdb.Gorm()
// 	db = db.Table("x_seller")
// 	err := db.Create(&reqdata).Error
// 	if err != nil {
// 		logs.Error(err)
// 		ctx.Put("error", err.Error())
// 		ctx.RespErr("添加失败")
// 		return
// 	}
// 	ctx.RespOK()
// }

// func modify_seller(ctx *XHttpContent) {
// 	type RequestData struct {
// 		SellerId   int    `validate:"required"`
// 		SellerName string `validate:"required"`
// 		State      int    `validate:"required"`
// 	}
// 	reqdata := RequestData{}
// 	if ctx.RequestData(&reqdata) != nil {
// 		return
// 	}
// 	db := thisdb.Gorm()
// 	db = db.Where("SellerId = ?", reqdata.SellerId)
// 	updatedata := *ObjectToMap(&reqdata)
// 	delete(updatedata, "SellerId")
// 	if reqdata.SellerName == "" {
// 		delete(updatedata, "SellerName")
// 	}
// 	if reqdata.State == 0 {
// 		delete(updatedata, "State")
// 	}
// 	err := db.Update(updatedata).Error
// 	if err != nil {
// 		logs.Error(err)
// 		ctx.Put("error", err.Error())
// 		ctx.RespErr("修改失败")
// 		return
// 	}
// 	ctx.RespOK()
// }

// func delete_seller(ctx *XHttpContent) {
// 	type RequestData struct {
// 		SellerId int `validate:"required"`
// 	}
// 	reqdata := RequestData{}
// 	if ctx.RequestData(&reqdata) != nil {
// 		return
// 	}
// 	db := thisdb.Gorm()
// 	db = db.Table("x_seller")
// 	db = db.Where("SellerId = ?", reqdata.SellerId)
// 	err := db.Delete(&reqdata).Error
// 	if err != nil {
// 		logs.Error(err)
// 		ctx.Put("error", err.Error())
// 		ctx.RespErr("删除失败")
// 		return
// 	}
// 	ctx.RespOK()
// }

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
	offset := (reqdata.Page - 1) * reqdata.PageSize
	channel := []XChannel{}
	db := thisdb.Gorm()
	db = db.Table("x_channel").Order("id desc")
	db = thisdb.Where(db, "SellerId = ?", reqdata.SellerId, int(0))
	db = thisdb.Where(db, "ChannelId = ?", reqdata.ChannelId, int(0))
	db = thisdb.Where(db, "ChannelName = ?", reqdata.ChannelName, "")
	var total int
	db.Count(&total)
	db.Offset(offset).Limit(reqdata.PageSize).Find(&channel)
	for i := 0; i < len(channel); i++ {
		channel[i].CreateTime = LocalTimeToUtc(channel[i].CreateTime)
	}
	ctx.Put("data", channel)
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
	db := thisdb.Gorm()
	db = db.Table("x_channel")
	err := db.Create(&reqdata).Error
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
	db := thisdb.Gorm()
	db = db.Table("x_channel")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("ChannelId = ?", reqdata.ChannelId)
	updatedata := *ObjectToMap(&reqdata)
	delete(updatedata, "SellerId")
	delete(updatedata, "ChannelId")
	err := db.Update(updatedata).Error
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
	db := thisdb.Gorm()
	db = db.Table("x_channel")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("ChannelId = ?", reqdata.ChannelId)
	err := db.Delete(&reqdata).Error
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
	if reqdata.Page <= 0 {
		reqdata.Page = 1
	}
	if reqdata.PageSize <= 0 || reqdata.PageSize > 2000 {
		reqdata.PageSize = 15
	}
	offset := (reqdata.Page - 1) * reqdata.PageSize
	role := []XAdminRole{}
	db := thisdb.Gorm()
	db = db.Table("x_admin_role").Order("id desc")
	db = thisdb.Where(db, "SellerId = ?", reqdata.SellerId, int(0))
	db = thisdb.Where(db, "RoleName = ?", reqdata.RoleName, "")
	var total int
	db.Count(&total)
	db.Offset(offset).Limit(reqdata.PageSize).Find(&role)
	for i := 0; i < len(role); i++ {
		role[i].CreateTime = LocalTimeToUtc(role[i].CreateTime)
	}
	ctx.Put("data", role)
	ctx.Put("total", total)
	ctx.RespOK()
}

func add_role(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int    `validate:"required" gorm:"column:SellerId"`
		RoleName string `validate:"required" gorm:"column:RoleName"`
		Parent   string `validate:"required" gorm:"column:Parent"`
		RoleData string `validate:"required" gorm:"column:RoleData"`
		Memo     string `gorm:"column:Memo"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	db := thisdb.Gorm()
	db = db.Table("x_admin_role")
	err := db.Create(&reqdata).Error
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
		RoleData string
		State    int
		Memo     string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	db := thisdb.Gorm()
	db = db.Table("x_admin_role")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("RoleName = ?", reqdata.RoleName)
	updatedata := *ObjectToMap(&reqdata)
	delete(updatedata, "SellerId")
	delete(updatedata, "RoleName")
	err := db.Update(updatedata).Error
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
	db := thisdb.Gorm()
	db = db.Table("x_admin_role")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("RoleName = ?", reqdata.RoleName)
	err := db.Delete(&reqdata).Error
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
	offset := (reqdata.Page - 1) * reqdata.PageSize
	user := []XAdminUser{}
	db := thisdb.Gorm()
	db = db.Table("x_admin_User").Order("id desc")
	db = thisdb.Where(db, "SellerId = ?", reqdata.SellerId, nil)
	db = thisdb.Where(db, "ChannelId = ?", reqdata.ChannelId, int(0))
	db = thisdb.Where(db, "Account = ?", reqdata.Account, "")
	var total int
	db.Count(&total)
	db.Offset(offset).Limit(reqdata.PageSize).Find(&user)
	for i := 0; i < len(user); i++ {
		user[i].CreateTime = LocalTimeToUtc(user[i].CreateTime)
		user[i].Token = ""
		user[i].Password = ""
		user[i].LoginGoogle = ""
		user[i].OptGoogle = ""
	}
	ctx.Put("data", user)
	ctx.Put("total", total)
	ctx.RespOK()
}

func add_admin_user(ctx *XHttpContent) {
	type RequestData struct {
		SellerId  int    `validate:"required" gorm:"column:SellerId"`
		ChannelId int    `gorm:"column:ChannelId"`
		Account   string `validate:"required" gorm:"column:Account"`
		Password  string `validate:"required" gorm:"column:Password"`
		RoleName  string `validate:"required" gorm:"column:RoleName"`
		Memo      string `gorm:"column:Memo"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	reqdata.Password = Md5(reqdata.Password)
	db := thisdb.Gorm()
	db = db.Table("x_admin_user")
	err := db.Create(&reqdata).Error
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
	db := thisdb.Gorm()
	db = db.Table("x_admin_user")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("Account = ?", reqdata.Account)
	updatedata := *ObjectToMap(&reqdata)
	delete(updatedata, "SellerId")
	delete(updatedata, "RoleName")
	if reqdata.Password == "" {
		delete(updatedata, "Password")
	}
	err := db.Update(updatedata).Error
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
	db := thisdb.Gorm()
	db = db.Table("x_admin_user")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = db.Where("Account = ?", reqdata.Account)
	err := db.Delete(&reqdata).Error
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

	user := XAdminUser{}
	err := thisdb.Gorm().Where("SellerId = ? and Account = ?", reqdata.SellerId, reqdata.Account).Table("x_admin_user").First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.RespErr("管理员不存在")
		return
	}

	me := XAdminUser{}
	thisdb.Gorm().Where("SellerId = ? and Account = ?", token.SellerId, token.Account).Table("x_admin_user").First(&me)

	if reqdata.Account != token.Account {
		if me.OptGoogle == "" {
			ctx.RespErr(fmt.Sprintf("请先设置账号 %v 的操作验证码", token.Account))
			return
		}
		if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(me.OptGoogle, reqdata.GoogleCode) {
			ctx.RespErr("验证码不正确")
			return
		}
	} else {
		if reqdata.CodeType == 2 && me.LoginGoogle == "" {
			if me.OptGoogle == "" {
				ctx.RespErr(fmt.Sprintf("请先设置账号 %v 的登录验证码", me.Account))
				return
			}
		}
		if me.OptGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(me.OptGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的操作验证码", me.Account))
				return
			}
		}
		if me.LoginGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(me.LoginGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的登录 验证码", me.Account))
				return
			}
		}
	}
	seller := XSeller{}
	thisdb.Gorm().Table("x_seller").Where("SellerId = ?", reqdata.SellerId).First(&seller)
	if reqdata.CodeType == 1 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-登录", reqdata.Account, verifykey, seller.SellerName)
		thisdb.Exec("update x_admin_user set LoginGoogle = ? where Id = ?", verifykey, user.Id)
		ctx.Put("url", verifyurl)
		ctx.RespOK()
	} else if reqdata.CodeType == 2 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-操作", reqdata.Account, verifykey, seller.SellerName)
		thisdb.Exec("update x_admin_user set OptGoogle = ? where Id = ?", verifykey, user.Id)
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
	if reqdata.Page <= 0 {
		reqdata.Page = 1
	}
	if reqdata.PageSize <= 0 {
		reqdata.PageSize = 15
	}
	if reqdata.StartTime != "" {
		reqdata.StartTime = UtcToLocalTime(reqdata.StartTime)
	}
	if reqdata.EndTime != "" {
		reqdata.EndTime = UtcToLocalTime(reqdata.EndTime)
	}
	db := thisdb.Gorm().Table("x_admin_login_log")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	db = thisdb.Where(db, "ChannelId = ?", reqdata.ChannelId, int(0))
	db = thisdb.Where(db, "Account = ?", reqdata.Account, "")
	db = thisdb.Where(db, "LoginIp = ?", reqdata.LoginIp, "")
	db = thisdb.Where(db, "CreateTime >= ?", reqdata.StartTime, "")
	db = thisdb.Where(db, "CreateTime < ?", reqdata.EndTime, "")
	offset := (reqdata.Page - 1) * reqdata.PageSize
	logs := []XAdminLoginLog{}
	db.Offset(offset).Limit(reqdata.PageSize).Find(&logs)
	logdata := []map[string]interface{}{}
	for i := 0; i < len(logs); i++ {
		mapdata := ObjectToMap(logs[i])
		(*mapdata)["IpLocation"] = GetIpLocation(logs[i].LoginIp)
		logdata = append(logdata, *mapdata)
	}
	ctx.Put("data", logdata)
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
	db := thisdb.Gorm().Table("x_config")
	db = db.Where("SellerId = ?", reqdata.SellerId)
	if reqdata.ChannelId == 0 {
		db = db.Where("ChannelId = ?", 0)
	} else {
		db = db.Where("ChannelId = 0 or ChannelId = ?", reqdata.ChannelId)
	}
	if len(reqdata.ConfigName) > 0 {
		db = db.Where("ConfigName in (?)", reqdata.ConfigName)
	}
	config := []XConfig{}
	db.Find(&config)
	for i := 0; i < len(config); i++ {
		config[i].CreateTime = LocalTimeToUtc(config[i].CreateTime)
	}
	ctx.Put("data", config)
	ctx.RespOK()
}

func modify_system_config(ctx *XHttpContent) {
	reqdata := ModifyConfigData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if beforeModifyConfig != nil {
		reqdata = beforeModifyConfig(reqdata)
	}
	for i := 0; i < len(reqdata.Config); i++ {
		sql := "update x_config set ConfigValue = ? where SellerId = ? and ChannelId = ? and ConfigName = ?"
		thisdb.Exec(sql, reqdata.Config[i].ConfigValue, reqdata.SellerId, reqdata.Config[i].ChannelId, reqdata.Config[i].ConfigName)
	}
	ctx.RespOK()
}
