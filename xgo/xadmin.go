package xgo

type XAdminRole struct {
	Id           uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId     int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	RoleName     string `gorm:"column:RoleName;index:RoleName;size:32;comment:'角色名'"`
	Parent       string `gorm:"column:Parent;size:32;comment:'上级角色'"`
	RoleData     string `gorm:"column:RoleData;type:text;comment:'权限数据'"`
	CreateTime   string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
	CreateTimeEx string `gorm:"column:CreateTimeEx;type:datetime;default:now();comment:'创建时间'"`
}

type XAdminUser struct {
	Id          uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId    int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	ChannelId   int    `gorm:"column:ChannelId;index:ChannelId;comment:'渠道商'"`
	Account     string `gorm:"column:Account;size:32;index:Account;comment:'账号'"`
	Password    string `gorm:"column:Password;size:64;comment:'登录密码'"`
	RoleName    string `gorm:"column:RoleName;size:32;comment:'角色'"`
	LoginGoogle string `gorm:"column:LoginGoogle;size:32;comment:'登录谷歌验证码'"`
	OptGoogle   string `gorm:"column:OptGoogle;size:32;comment:'渠道商'"`
	State       int    `gorm:"column:State;comment:'状态 1开启,2关闭'"`
	Token       string `gorm:"column:Token;size:64;comment:'最后一次登录的token'"`
	LoginTime   int    `gorm:"column:LoginTime;comment:'登录次数'"`
	LoginIp     string `gorm:"column:LoginIp;size:32;comment:'最近一次登录Ip'"`
	CreateTime  string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

type XAdminLoginLog struct {
	Id         uint   `gorm:"column:Id;primary_key;auto_increment;comment:'自增Id'"`
	SellerId   int    `gorm:"column:SellerId;index:SellerId;comment:'运营商'"`
	ChannelId  int    `gorm:"column:ChannelId;index:ChannelId;comment:'渠道商'"`
	Account    string `gorm:"column:Account;size:32;index:Account;comment:'账号'"`
	Token      string `gorm:"column:Token;size:64;comment:'登录的token'"`
	LoginIp    string `gorm:"column:LoginIp;size:32;comment:'最近一次登录Ip'"`
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
	CreateTime string `gorm:"column:CreateTime;type:datetime;default:now();comment:'创建时间'"`
}

func AdminInit(http *XHttp, db *XDb) {
	db.Gorm().Table("x_admin_role").AutoMigrate(&XAdminRole{})
	db.Gorm().Table("x_admin_user").AutoMigrate(&XAdminUser{})
	db.Gorm().Table("x_admin_login_log").AutoMigrate(&XAdminLoginLog{})
	db.Gorm().Table("x_admin_opt_log").AutoMigrate(&XAdminOptLog{})
}
