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

type AdminTokenData struct {
	Account   string
	SellerId  int
	ChannelId int
	AuthData  string
}

type XAdminRole struct {
	Id         uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId   int    `gorm:"column:SellerId;UNIQUE_INDEX:idx_sr;index:SellerId;comment:'运营商'"`
	RoleName   string `gorm:"column:RoleName;UNIQUE_INDEX:idx_sr;index:RoleName;size:32;comment:'角色名'"`
	Parent     string `gorm:"column:Parent;size:32;comment:'上级角色'"`
	RoleData   string `gorm:"column:RoleData;type:text;comment:'权限数据'"`
	State      int    `gorm:"column:State;default:1;comment:'状态 1开启,2关闭'"`
	Memo       string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XAdminUser struct {
	Id          uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId    int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	ChannelId   int    `gorm:"column:ChannelId;index:ChannelId;default:0;comment:'渠道商'"`
	Account     string `gorm:"column:Account;size:32;index:Account;comment:'账号'"`
	Password    string `gorm:"column:Password;size:64;comment:'登录密码'"`
	RoleName    string `gorm:"column:RoleName;size:32;comment:'角色'"`
	LoginGoogle string `gorm:"column:LoginGoogle;size:32;comment:'登录谷歌验证码'"`
	OptGoogle   string `gorm:"column:OptGoogle;size:32;comment:'渠道商'"`
	State       int    `gorm:"column:State;default:1;comment:'状态 1开启,2关闭'"`
	Token       string `gorm:"column:Token;size:64;comment:'最后登录的token'"`
	LoginCount  int    `gorm:"column:LoginCount;default:0;comment:'登录次数'"`
	LoginTime   string `gorm:"column:LoginTime;type:datetime;default:now();comment:'最后登录时间'"`
	LoginIp     string `gorm:"column:LoginIp;size:32;comment:'最后登录Ip'"`
	Memo        string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime  string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XAdminLoginLog struct {
	Id         uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId   int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	ChannelId  int    `gorm:"column:ChannelId;index:ChannelId;comment:'渠道商'"`
	Account    string `gorm:"column:Account;size:32;index:Account;comment:'账号'"`
	Token      string `gorm:"column:Token;size:64;comment:'登录的token'"`
	LoginIp    string `gorm:"column:LoginIp;size:32;comment:'最近一次登录Ip'"`
	Memo       string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XAdminOptLog struct {
	Id         uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId   int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	ChannelId  int    `gorm:"column:ChannelId;index:ChannelId;comment:'渠道商'"`
	Account    string `gorm:"column:Account;size:32;index:Account;comment:'账号'"`
	ReqPath    string `gorm:"column:ReqPath;size:256;comment:'请求路径'"`
	ReqData    string `gorm:"column:ReqData;size:256;comment:'请求参数'"`
	Ip         string `gorm:"column:Ip;size:32;comment:'请求的Ip'"`
	Memo       string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XSeller struct {
	Id         uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId   int    `gorm:"column:SellerId;index:SellerId;unique;comment:'运营商'"`
	State      int    `gorm:"column:State;default:1;comment:'状态 1开启,2关闭'"`
	SellerName string `gorm:"column:SellerName;size:32;comment:'运营商名称'"`
	Memo       string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XChannel struct {
	Id          uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId    int    `gorm:"column:SellerId;index:SellerId;UNIQUE_INDEX:idx_sc;comment:'运营商'"`
	ChannelId   int    `gorm:"column:ChannelId;index:ChannelId;UNIQUE_INDEX:idx_sc;comment:'渠道商'"`
	State       int    `gorm:"column:State;default:1;comment:'状态 1开启,2关闭'"`
	ChannelName string `gorm:"column:ChannelName;size:32;comment:'渠道名称'"`
	Memo        string `gorm:"column:Memo;size:256;comment:'备注'"`
	CreateTime  string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

func AdminInit(http *XHttp, db *XDb, redis *XRedis, fullauth string) {
	thishttp = http
	thisdb = db
	thisredis = redis
	if env != "dev" {
		db.Gorm().Table("x_admin_role").AutoMigrate(&XAdminRole{})
		db.Gorm().Table("x_admin_user").AutoMigrate(&XAdminUser{})
		db.Gorm().Table("x_admin_login_log").AutoMigrate(&XAdminLoginLog{})
		db.Gorm().Table("x_admin_opt_log").AutoMigrate(&XAdminOptLog{})
		db.Gorm().Table("x_seller").AutoMigrate(&XSeller{})
		db.Gorm().Table("x_channel").AutoMigrate(&XChannel{})
		var count int
		db.Gorm().Table("x_seller").Count(&count)
		if count == 0 {
			seller := XSeller{}
			seller.SellerId = 1
			seller.SellerName = "初始运营商"
			db.Gorm().Create(&seller)
		}
		db.Gorm().Table("x_channel").Count(&count)
		if count == 0 {
			channel := XChannel{}
			channel.SellerId = 1
			channel.ChannelId = 1
			channel.ChannelName = "初始渠道"
			db.Gorm().Create(&channel)
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

	http.OnPostWithAuth("/sapi/get_login_log", get_login_log, "系统管理.登录日志.查", false, "")
	http.OnPostWithAuth("/sapi/get_opt_log", get_opt_log, "系统管理.操作日志.查", false, "")
}

func GetAdminToken(token string) *AdminTokenData {
	return nil
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
	thisdb.Gorm().Find(&sellers)
	for i := 0; i < len(sellers); i++ {
		role := XAdminRole{}
		err := thisdb.Gorm().Where("SellerId = ? and RoleName = '运营商超管'", sellers[i].SellerId).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role.SellerId = sellers[i].SellerId
			role.Parent = "god"
			role.RoleName = "运营商超管"
			role.RoleData = authstr
			thisdb.Gorm().Create(&role)

		}
		user := XAdminUser{}
		err = thisdb.Gorm().Where("SellerId = ?", sellers[i].SellerId).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.SellerId = sellers[i].SellerId
			user.Account = fmt.Sprintf("admin%v", user.SellerId)
			user.Password = Md5(Md5("admin"))
			user.RoleName = "运营商超管"
			thisdb.Gorm().Create(&user)

		}
	}
	sql := "update x_admin_role set RoleData = ? where RoleName = ?"
	db.conn().Exec(sql, authstr, "运营商超管")

	super := XAdminRole{}
	err := thisdb.Gorm().Where("SellerId = -1 and RoleName = '超级管理员'").First(&super).Error
	if super.RoleData != fullauth {
		roles := []XAdminRole{}
		thisdb.Gorm().Find(&roles)
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
		thisdb.Gorm().Create(&super)
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

func get_login_log(ctx *XHttpContent) {
}

func get_opt_log(ctx *XHttpContent) {
}
