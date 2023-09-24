package xgo

import (
	"encoding/json"
	"fmt"
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
	ForClient   int
	Memo        string
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
	http.OnPostWithAuth("/sapi/add_system_config", add_system_config, "系统管理.系统设置.增", false, "新增系统设置")
	http.OnPostWithAuth("/sapi/modify_system_config", modify_system_config, "系统管理.系统设置.改", false, "修改系统设置")
}

func GetAdminToken(ctx *XHttpContent) *AdminTokenData {
	tokendata := AdminTokenData{}
	err := json.Unmarshal([]byte(ctx.TokenData), &tokendata)
	if err != nil {
		return nil
	}
	return &tokendata
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

	if !thisredis.GetLock(fmt.Sprintf("lock:admin_login:%v", reqdata.Account), 10) {
		ctx.RespErr("操作频繁,请稍后再试")
		return
	}

	user, err := thisdb.Table("x_admin_user").Where("Account = ?", reqdata.Account, nil).First()
	if err != nil {
		logs.Error("user_login:", err)
		ctx.RespErr("登录失败,请稍后再试")
	}
	if user == nil {
		ctx.RespErr("账号不存在")
		return
	}
	if user.Int("State") != 1 {
		ctx.RespErr("账号已禁用")
		return
	}
	if strings.Index(env, "prd") >= 0 && user.String("LoginGoogle") != "" && !VerifyGoogleCode(user.String("LoginGoogle"), reqdata.GoogleCode) {
		ctx.RespErr("验证码错误")
		return
	}
	if user.String("Password") != reqdata.Password {
		ctx.RespErr("密码不正确")
		return
	}

	seller, err := thisdb.Table("x_seller").Where("SellerId = ?", user.String("SellerId"), nil).First()
	if err != nil {
		logs.Error("user_login:", err)
		ctx.RespErr("登录失败,请稍后再试")
		return
	}
	if seller == nil {
		ctx.RespErr("运营商不存在")
		return
	}
	if seller.Int("State") != 1 {
		ctx.RespErr("运营商已禁用")
		return
	}
	role, err := thisdb.Table("x_admin_role").Where("SellerId = ?", user.Int("SellerId"), nil).
		Where("RoleName = ?", user.String("RoleName"), nil).First()
	if err != nil {
		logs.Error("user_login:", err)
		ctx.RespErr("登录失败,请稍后再试")
	}
	if role == nil {
		ctx.RespErr("角色不存在")
		return
	}
	if role.Int("State") != 1 {
		ctx.RespErr("角色已禁用")
		return
	}
	if len(user.String("Token")) > 0 {
		thishttp.DelToken(user.String("Token"))
	}
	if user.Int("ChannelId") != 0 {
		ctx.RespErr("该账号为渠道账号,不可登录运营后台")
		return
	}
	token := uuid.New().String()
	tokendata := AdminTokenData{}
	tokendata.Account = reqdata.Account
	tokendata.SellerId = user.Int("SellerId")
	tokendata.ChannelId = user.Int("ChannelId")
	tokendata.AuthData = role.String("RoleData")
	thishttp.SetToken(token, tokendata)
	sql := "update x_admin_user set Token = ?,LoginCount = LoginCount + 1,LoginTime = now(),LoginIp = ? where id = ?"
	thisdb.Exec(sql, token, ctx.GetIp(), user.Int("Id"))
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
	ctx.Put("UserId", user.Int("Id"))
	ctx.Put("SellerId", user.Int("SellerId"))
	ctx.Put("ChannelId", user.Int("ChannelId"))
	ctx.Put("Account", reqdata.Account)
	ctx.Put("Token", token)
	ctx.Put("LoginTime", user.String("LoginTime"))
	ctx.Put("Ip", ctx.GetIp())
	ctx.Put("LoginCount", user.Int("LoginCount"))
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
	sellers, err := thisdb.Table("x_seller").Select("SellerId,SellerName").Find()
	if err != nil {
		ctx.RespErr(err)
		return
	}
	ctx.RespOK(sellers.Maps())
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
	channels, err := table.Find()
	if err != nil {
		ctx.RespErr(err)
		return
	}
	ctx.RespOK(channels.Maps())
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
	roles, err := table.Find()
	if err != nil {
		ctx.RespErr(err)
		return
	}
	ctx.Put("data", roles.Maps())
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
		role, err := table.First()
		if err != nil {
			ctx.RespErr(err)
			return
		}
		if role != nil {
			ctx.Put("RoleData", role.String("RoleData"))
		}
	}
	{
		table := thisdb.Table("x_admin_role")
		table = table.Where("SellerId = ?", reqdata.SellerId, nil)
		table = table.Where("RoleName = ?", "运营商超管", nil)
		table = table.Select("RoleData")
		role, err := table.First()
		if err != nil {
			ctx.RespErr(err)
			return
		}
		if role != nil {
			ctx.Put("SuperRoleData", role.String("RoleData"))
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
	total, err := table.Count()
	if err != nil {
		ctx.RespErr(err)
		return
	}
	channels, err := table.Find()
	if err != nil {
		ctx.RespErr(err)
		return
	}
	ctx.Put("data", channels.Maps())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.RespOK()
}

func get_role(ctx *XHttpContent) {
	type RequestData struct {
		SellerId int `validate:"required"`
		Page     int
		PageSize int
		RoleName string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_admin_role").OrderBy("id desc")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("RoleName = ?", reqdata.RoleName, "")
	total, err := table.Count()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	roles, err := table.PageData(reqdata.Page, reqdata.PageSize)
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.Put("data", roles.Maps())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
	total, err := table.Count()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	users, err := table.Select("*").PageData(reqdata.Page, reqdata.PageSize)
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	users.ForEach(func(xd *XMap) bool {
		xd.Delete("Token")
		xd.Delete("Password")
		xd.Delete("LoginGoogle")
		xd.Delete("OptGoogle")
		return true
	})
	ctx.Put("data", users.Maps())
	ctx.Put("total", total)
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
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
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.RespOK()
}

func modify_admin_user_google(ctx *XHttpContent) {
	type RequestData struct {
		SellerId   int    `validate:"required"`
		Account    string `validate:"required"`
		CodeType   int    `validate:"required"`
		GoogleCode string `validate:"required"`
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	token := GetAdminToken(ctx)

	if !thisredis.GetLock(fmt.Sprintf("lock:admin_change_google:%v", reqdata.Account), 10) {
		ctx.RespErr("操作频繁,请稍后再试")
		return
	}

	user, err := thisdb.Table("x_admin_user").Where("SellerId = ?", reqdata.SellerId, nil).Where("Account = ?", reqdata.Account, nil).First()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	if user == nil {
		ctx.RespErr("管理员不存在")
		return
	}

	me, err := thisdb.Table("x_admin_user").Where("SellerId = ?", token.SellerId, nil).Where("Account = ?", token.Account, nil).First()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	OptGoogle := me.String("OptGoogle")
	LoginGoogle := me.String("LoginGoogle")
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
				ctx.RespErr(fmt.Sprintf("请先设置账号 %v 的登录验证码", me.String("Account")))
				return
			}
		}
		if OptGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(OptGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的操作验证码", me.String("Account")))
				return
			}
		}
		if LoginGoogle != "" {
			if strings.Index(env, "prd") >= 0 && VerifyGoogleCode(LoginGoogle, reqdata.GoogleCode) {
				ctx.RespErr(fmt.Sprintf("验证码不正确,请输入账号 %v 的登录 验证码", me.String("Account")))
				return
			}
		}
	}
	seller, err := thisdb.Table("x_seller").Where("SellerId = ?", reqdata.SellerId, nil).First()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	if seller == nil {
		ctx.RespErr("运营商不存在")
		return
	}
	if reqdata.CodeType == 1 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-登录", reqdata.Account, verifykey, seller.String("SellerName"))
		thisdb.Exec("update x_admin_user set LoginGoogle = ? where Id = ?", verifykey, user.Int("Id"))
		ctx.Put("url", verifyurl)
		ctx.RespOK()
	} else if reqdata.CodeType == 2 {
		verifykey := NewGoogleSecret()
		verifyurl := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s-操作", reqdata.Account, verifykey, seller.String("SellerName"))
		thisdb.Exec("update x_admin_user set OptGoogle = ? where Id = ?", verifykey, user.Int("Id"))
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
	total, err := table.Count()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	data, err := table.Select("*").PageData(reqdata.Page, reqdata.PageSize)
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.Put("data", data.Maps())
	ctx.Put("total", total)
	ctx.RespOK()
}

func get_opt_log(ctx *XHttpContent) {
	type RequestData struct {
		Page      int
		PageSize  int
		SellerId  int `validate:"required" `
		ChannelId int
		Account   string
		OptName   string
		Ip        string
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
	table := thisdb.Table("x_admin_opt_log")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, 0)
	table = table.Where("Account = ?", reqdata.Account, "")
	table = table.Where("OptName = ?", reqdata.OptName, "")
	table = table.Where("Ip = ?", reqdata.Ip, "")
	table = table.Where("CreateTime >= ?", reqdata.StartTime, "")
	table = table.Where("CreateTime < ?", reqdata.EndTime, "")
	total, err := table.Count()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	data, err := table.PageData(reqdata.Page, reqdata.PageSize)
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.Put("data", data.Maps())
	ctx.Put("total", total)
	ctx.RespOK()
}

func get_system_config(ctx *XHttpContent) {
	type RequestData struct {
		SellerId   int `validate:"required" `
		ChannelId  int
		ConfigName []interface{}
		Memo       string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	table := thisdb.Table("x_config")
	table = table.Where("SellerId = ?", reqdata.SellerId, nil)
	table = table.Where("ChannelId = ?", reqdata.ChannelId, -1)
	table = table.Where("ConfigName in ", reqdata.ConfigName)
	if reqdata.Memo != "" {
		table = table.Where("Memo like ?", fmt.Sprintf("%%%v%%", reqdata.Memo), nil)
	}
	total, err := table.Count()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	config, err := table.Find()
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
	ctx.Put("data", config.Maps())
	ctx.Put("total", total)
	ctx.RespOK()
}

func add_system_config(ctx *XHttpContent) {
	type RequestData struct {
		SellerId    int `validate:"required" `
		ChannelId   int
		ConfigName  string
		ConfigValue string
		ForClient   int
		Memo        string
	}
	reqdata := RequestData{}
	if ctx.RequestData(&reqdata) != nil {
		return
	}
	if reqdata.ForClient != 1 {
		reqdata.ForClient = 2
	}
	_, err := thisdb.Table("x_config").Insert(reqdata)
	if err != nil {
		logs.Error(err.Error())
		ctx.RespErr(err.Error())
		return
	}
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
		if reqdata.Config[i].Memo != "" {
			sql := "update x_config set ConfigValue = ?,Memo =?,ForClient = ? where SellerId = ? and ChannelId = ? and ConfigName = ?"
			thisdb.Exec(sql, reqdata.Config[i].ConfigValue, reqdata.Config[i].Memo, reqdata.Config[i].ForClient, reqdata.SellerId, reqdata.Config[i].ChannelId, reqdata.Config[i].ConfigName)
		} else {
			sql := "update x_config set ConfigValue = ?,ForClient = ? where SellerId = ? and ChannelId = ? and ConfigName = ?"
			thisdb.Exec(sql, reqdata.Config[i].ConfigValue, reqdata.Config[i].ForClient, reqdata.SellerId, reqdata.Config[i].ChannelId, reqdata.Config[i].ConfigName)
		}
	}
	ctx.RespOK()
}
