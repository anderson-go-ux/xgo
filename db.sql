CREATE TABLE IF NOT EXISTS  x_admin_login_log (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  Account varchar(32)  DEFAULT NULL COMMENT '账号',
  Token varchar(64)  DEFAULT NULL COMMENT '登录的token',
  LoginIp varchar(32)  DEFAULT NULL COMMENT '最近一次登录Ip',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  KEY SellerId (SellerId),
  KEY ChannelId (ChannelId),
  KEY Account (Account),
  KEY LoginIp (LoginIp)
) ENGINE=InnoDB AUTO_INCREMENT=0  ;

CREATE TABLE IF NOT EXISTS  x_admin_opt_log (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  Account varchar(32)  DEFAULT NULL COMMENT '账号',
  ReqPath varchar(256)  DEFAULT NULL COMMENT '请求路径',
  ReqData varchar(256)  DEFAULT NULL COMMENT '请求参数',
  Ip varchar(32)  DEFAULT NULL COMMENT '请求的Ip',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  KEY SellerId (SellerId),
  KEY ChannelId (ChannelId),
  KEY Account (Account)
) ENGINE=InnoDB ;

CREATE TABLE IF NOT EXISTS  x_admin_role (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  RoleName varchar(32)  DEFAULT NULL COMMENT '角色名',
  Parent varchar(32)  DEFAULT NULL COMMENT '上级角色',
  RoleData text  COMMENT '权限数据',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  UNIQUE KEY idx_sr (SellerId,RoleName),
  KEY SellerId (SellerId),
  KEY RoleName (RoleName)
) ENGINE=InnoDB AUTO_INCREMENT=0  ;

CREATE TABLE IF NOT EXISTS  x_admin_user (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT '0' COMMENT '渠道商',
  Account varchar(32)  DEFAULT NULL COMMENT '账号',
  Password varchar(64)  DEFAULT NULL COMMENT '登录密码',
  RoleName varchar(32)  DEFAULT NULL COMMENT '角色',
  LoginGoogle varchar(32)  DEFAULT NULL COMMENT '登录谷歌验证码',
  OptGoogle varchar(32)  DEFAULT NULL COMMENT '渠道商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  Token varchar(64)  DEFAULT NULL COMMENT '最后登录的token',
  LoginCount int DEFAULT '0' COMMENT '登录次数',
  LoginTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '最后登录时间',
  LoginIp varchar(32)  DEFAULT NULL COMMENT '最后登录Ip',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  KEY SellerId (SellerId),
  KEY ChannelId (ChannelId),
  KEY Account (Account)
) ENGINE=InnoDB AUTO_INCREMENT=0  ;

CREATE TABLE IF NOT EXISTS  x_channel (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  ChannelName varchar(32)  DEFAULT NULL COMMENT '渠道名称',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  UNIQUE KEY idx_sc (SellerId,ChannelId),
  KEY SellerId (SellerId),
  KEY ChannelId (ChannelId)
) ENGINE=InnoDB AUTO_INCREMENT=0  ;

CREATE TABLE IF NOT EXISTS  x_config (
  Id bigint unsigned NOT NULL AUTO_INCREMENT,
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT '0' COMMENT '渠道',
  ConfigName varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '配置名',
  ConfigValue text CHARACTER SET utf8mb4  COMMENT '配置值',
  EditAble int DEFAULT '1' COMMENT '是否可编辑',
  ShowAble int DEFAULT '1' COMMENT '是否在后台显示',
  ForClient int DEFAULT '2' COMMENT '该配置客户端是否能获取',
  Memo varchar(1024)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  KEY SellerId (SellerId),
  KEY ChannelId (ChannelId),
  KEY ConfigName (ConfigName)
) ENGINE=InnoDB ;

CREATE TABLE IF NOT EXISTS  x_seller (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  SellerName varchar(32)  DEFAULT NULL COMMENT '运营商名称',
  Memo varchar(256)  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id),
  UNIQUE KEY SellerId (SellerId)
) ENGINE=InnoDB AUTO_INCREMENT=0  ;

/*
type x_admin_login_log struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	Account string `gorm:"column:Account"`
	Token string `gorm:"column:Token"`
	LoginIp string `gorm:"column:LoginIp"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_admin_opt_log struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	Account string `gorm:"column:Account"`
	ReqPath string `gorm:"column:ReqPath"`
	ReqData string `gorm:"column:ReqData"`
	Ip string `gorm:"column:Ip"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_admin_role struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	RoleName string `gorm:"column:RoleName"`
	Parent string `gorm:"column:Parent"`
	RoleData string `gorm:"column:RoleData"`
	State int `gorm:"column:State"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_admin_user struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	Account string `gorm:"column:Account"`
	Password string `gorm:"column:Password"`
	RoleName string `gorm:"column:RoleName"`
	LoginGoogle string `gorm:"column:LoginGoogle"`
	OptGoogle string `gorm:"column:OptGoogle"`
	State int `gorm:"column:State"`
	Token string `gorm:"column:Token"`
	LoginCount int `gorm:"column:LoginCount"`
	LoginTime string `gorm:"column:LoginTime"`
	LoginIp string `gorm:"column:LoginIp"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_channel struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	State int `gorm:"column:State"`
	ChannelName string `gorm:"column:ChannelName"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_config struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	ConfigName string `gorm:"column:ConfigName"`
	ConfigValue string `gorm:"column:ConfigValue"`
	EditAble int `gorm:"column:EditAble"`
	ShowAble int `gorm:"column:ShowAble"`
	ForClient int `gorm:"column:ForClient"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_seller struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	State int `gorm:"column:State"`
	SellerName string `gorm:"column:SellerName"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

*/
