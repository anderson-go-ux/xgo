CREATE TABLE IF NOT EXISTS  _muban (
  Id bigint NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道',
  UserId int DEFAULT NULL COMMENT '玩家Id',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP,
  Abc varchar(255)  DEFAULT NULL,
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY UserId (UserId) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_admin_login_log (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  Account varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '账号',
  Token varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '登录的token',
  LoginIp varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '最近一次登录Ip',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY Account (Account) USING BTREE,
  KEY LoginIp (LoginIp) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_admin_opt_log (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  Account varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '账号',
  OptName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL,
  ReqPath varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '请求路径',
  ReqData varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '请求参数',
  Ip varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '请求的Ip',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY Account (Account) USING BTREE,
  KEY OptName (OptName) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_admin_role (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  RoleName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '角色名',
  Parent varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '上级角色',
  RoleData text CHARACTER SET utf8mb4  COMMENT '权限数据',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY idx_sr (SellerId,RoleName) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY RoleName (RoleName) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_admin_user (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT '0' COMMENT '渠道商',
  Account varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '账号',
  Password varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '登录密码',
  RoleName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '角色',
  LoginGoogle varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '登录谷歌验证码',
  OptGoogle varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '渠道商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  Token varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '最后登录的token',
  LoginCount int DEFAULT '0' COMMENT '登录次数',
  LoginTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '最后登录时间',
  LoginIp varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '最后登录Ip',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY Account (Account) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_agent (
  Id bigint NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道',
  UserId int DEFAULT NULL COMMENT '玩家Id',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY UserId (UserId) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_agent_child (
  Id bigint NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道',
  UserId int DEFAULT NULL COMMENT '玩家Id',
  ChildId int DEFAULT NULL COMMENT '下级Id',
  ChildLevel int DEFAULT NULL COMMENT '第几层下级 0是直属下级',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '生成关系时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY UserChild (UserId,ChildId) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY UserId (UserId) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_channel (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  ChannelName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '渠道名称',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY idx_sc (SellerId,ChannelId) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_config (
  Id bigint unsigned NOT NULL AUTO_INCREMENT,
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT '0' COMMENT '渠道',
  ConfigName varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '配置名',
  ConfigValue text CHARACTER SET utf8mb4  COMMENT '配置值',
  ForClient int DEFAULT '2' COMMENT '该配置客户端是否能获取',
  Memo varchar(1024) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY ConfigName (ConfigName) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_error (
  Id bigint NOT NULL AUTO_INCREMENT,
  FunName varchar(255)  NOT NULL,
  ErrCode int NOT NULL,
  ErrMsg varchar(1024)  NOT NULL,
  CreateTime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (Id) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_seller (
  Id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  State int DEFAULT '1' COMMENT '状态 1开启,2关闭',
  SellerName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '运营商名称',
  Memo varchar(256) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '备注',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY SellerId (SellerId) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_user (
  Id bigint unsigned NOT NULL AUTO_INCREMENT,
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道',
  UserId bigint DEFAULT NULL COMMENT '玩家id',
  State int DEFAULT '1' COMMENT '状态 1启用,2禁用',
  Account varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '账号',
  Password varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '密码',
  Token varchar(64) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '最后登录token',
  NickName varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '昵称',
  PhoneNum varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '电话号码',
  Email varchar(255) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT 'Email地址',
  TopAgent int DEFAULT NULL COMMENT '顶级代理',
  Agents text CHARACTER SET utf8mb4  COMMENT '代理',
  Agent int DEFAULT NULL COMMENT '上级代理',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY UserId (UserId) USING BTREE,
  UNIQUE KEY SellerChannelAccount (SellerId,ChannelId,Account) USING BTREE,
  UNIQUE KEY PhoneNum (PhoneNum) USING BTREE,
  UNIQUE KEY Email (Email) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_user_pool (
  UserId int NOT NULL COMMENT '玩家Id',
  State int DEFAULT '1' COMMENT '状态',
  Ip varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '注册ip',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  PRIMARY KEY (UserId) USING BTREE,
  KEY State (State) USING BTREE,
  KEY CreateTime (CreateTime) USING BTREE
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS  x_user_score (
  Id bigint NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  SellerId int DEFAULT NULL COMMENT '运营商',
  ChannelId int DEFAULT NULL COMMENT '渠道',
  UserId int DEFAULT NULL COMMENT '玩家Id',
  Symbol varchar(32) CHARACTER SET utf8mb4  DEFAULT NULL COMMENT '币种',
  Amount decimal(50,6) DEFAULT '0.000000' COMMENT '可用金额',
  FrozenAmount decimal(50,6) DEFAULT '0.000000' COMMENT '冻结金额',
  CreateTime datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (Id) USING BTREE,
  UNIQUE KEY UserSymbol (UserId,Symbol) USING BTREE,
  KEY UserId (UserId) USING BTREE,
  KEY SellerId (SellerId) USING BTREE,
  KEY ChannelId (ChannelId) USING BTREE,
  KEY Symbol (Symbol) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 ROW_FORMAT=DYNAMIC;

DROP PROCEDURE IF EXISTS `x_init_auth`;
delimiter ;;
CREATE PROCEDURE `x_init_auth`()
BEGIN
	DECLARE p_done INT DEFAULT 0;
	DECLARE p_sellerid INT DEFAULT 0;
	DECLARE p_Id INT DEFAULT 0;
	DECLARE p_roledata text DEFAULT '{}';
	
	DECLARE cursor_seller CURSOR FOR SELECT SellerId FROM x_seller;
	DECLARE cursor_role CURSOR FOR SELECT Id,RoleData FROM x_admin_role WHERE RoleName <> '超级管理员' AND RoleName <> '运营商超管';
	
	DECLARE CONTINUE HANDLER FOR NOT FOUND SET p_done = 1;
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		GET CURRENT DIAGNOSTICS CONDITION 1	@errcode = MYSQL_ERRNO, @errmsg = MESSAGE_TEXT;
		ROLLBACK;
		INSERT INTO x_error(FunName,ErrCode,ErrMsg)VALUES('x_init_auth',@errcode,@errmsg);
		SELECT @errcode AS errcode,@errmsg AS errmsg;
	END;
	
	SET @fullauth = 
	'{
		"系统首页": { "查" : 1                              },
		"系统管理": {
			"系统设置":   { "查": 1,"增": 1,"改": 1         },
			"渠道管理":   { "查": 1,"增": 1,"删": 1,"改": 1 },
			"账号管理":   { "查": 1,"增": 1,"删": 1,"改": 1 },
			"角色管理":   { "查": 1,"增": 1,"删": 1,"改": 1 },
			"登录日志":   { "查": 1                         },
			"操作日志":   { "查": 1                         }
		}
	}';
	
	UPDATE x_admin_role SET RoleData = @fullauth WHERE RoleName = '超级管理员' OR RoleName = '运营商超管';
	
	IF NOT EXISTS(SELECT * FROM x_admin_role WHERE RoleName = '超级管理员') THEN
		INSERT INTO x_admin_role(SellerId,Parent,RoleName,RoleData)VALUES(-1,'god','超级管理员',@fullauth);
	END IF;
	
	OPEN cursor_seller;
    seller_loop: LOOP
		SET p_done = 0;
        FETCH cursor_seller INTO p_sellerid;
        IF p_done THEN
            LEAVE seller_loop;
        END IF;
		IF NOT EXISTS(SELECT * FROM x_admin_role WHERE SellerId = p_sellerid AND RoleName = '运营商超管') THEN
			INSERT INTO x_admin_role(SellerId,Parent,RoleName,RoleData)VALUES(p_sellerid,'god','运营商超管',@fullauth);
		END IF;
		
		IF NOT EXISTS(SELECT * FROM x_admin_user WHERE SellerId = p_sellerid) THEN
			INSERT INTO x_admin_user(SellerId,Account,`Password`,RoleName)VALUES(p_sellerid,CONCAT('admin',p_sellerid),MD5(MD5('admin')),'运营商超管');
		END IF;
    END LOOP;
    CLOSE cursor_seller;
	
	SET @tmpauth = '{}';
	SET @authkeys = JSON_KEYS(@fullauth);
	SET @idx = 0;
	WHILE @idx < JSON_LENGTH(@authkeys) DO
		SET @keyname = JSON_EXTRACT(@authkeys, CONCAT('$[',@idx,']'));
		SET @sub = JSON_EXTRACT(@fullauth,CONCAT('$.',@keyname));
		SET @sub =  JSON_UNQUOTE(@sub);
		SET @subkeys = JSON_KEYS(@sub);
		SET @subidx = 0;
		WHILE @subidx < JSON_LENGTH(@subkeys) DO
			SET @subkeyname = JSON_EXTRACT(@subkeys, CONCAT('$[',@subidx,']'));
			set @val = JSON_EXTRACT(@sub,CONCAT('$.',@subkeyname));
			SET @tmpauth = JSON_SET(@tmpauth, CONCAT('$."',JSON_UNQUOTE(@keyname),'.',JSON_UNQUOTE(@subkeyname),'"'), @val);
			SET @subidx = @subidx + 1;
		END WHILE;
		SET @idx = @idx + 1;
	END WHILE;
	
	SET @finalauth = '{}';
	SET @authkeys = JSON_KEYS(@tmpauth);
	SET @idx = 0;
	WHILE @idx < JSON_LENGTH(@authkeys) DO
		SET @keyname = JSON_EXTRACT(@authkeys, CONCAT('$[',@idx,']'));
		SET @sub = JSON_EXTRACT(@tmpauth,CONCAT('$.',@keyname));
		SET @sub =  JSON_UNQUOTE(@sub);
		IF @sub = '0' OR @sub = '1' THEN
			SET @finalauth = JSON_SET(@finalauth, CONCAT('$.',@keyname), CAST(@sub AS UNSIGNED));
		ELSE
			SET @subkeys = JSON_KEYS(@sub);
			SET @subidx = 0;
			WHILE @subidx < JSON_LENGTH(@subkeys) DO
				SET @subkeyname = JSON_EXTRACT(@subkeys, CONCAT('$[',@subidx,']'));
				set @val = JSON_EXTRACT(@sub,CONCAT('$.',@subkeyname));
				SET @val =  JSON_UNQUOTE(@val);
				SET @finalauth = JSON_SET(@finalauth, CONCAT('$."',JSON_UNQUOTE(@keyname),'.',JSON_UNQUOTE(@subkeyname),'"'), CAST(@val AS SIGNED));
				SET @subidx = @subidx + 1;
			END WHILE;
		END IF;
		SET @idx = @idx + 1;
	END WHILE;
	
	SET @authkeys = JSON_KEYS(@finalauth);
	
	OPEN cursor_role;
    role_loop: LOOP
		SET p_done = 0;
        FETCH cursor_role INTO p_Id,p_roledata;
        IF p_done THEN
            LEAVE role_loop;
        END IF;
	
		SET @idx = 0;
		WHILE @idx < JSON_LENGTH(@authkeys) DO
			SET @keyname = JSON_EXTRACT(@authkeys, CONCAT('$[',@idx,']'));
			SET @superval = JSON_EXTRACT(@finalauth,CONCAT('$.',@keyname));
			SET @keyname = REPLACE(@keyname,'.','"."');
			SET @keyname = CONCAT('$.',@keyname);
			IF @superval = '0' THEN
				SET p_roledata = JSON_SET(p_roledata, @keyname,0);
			END IF;
			SET @idx = @idx + 1;
		END WHILE;
		UPDATE x_admin_role SET RoleData = p_roledata where Id = p_Id;
    END LOOP;
    CLOSE cursor_role;
END
;;
delimiter ;

/*
type _muban struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	UserId int `gorm:"column:UserId"`
	CreateTime string `gorm:"column:CreateTime"`
	Abc string `gorm:"column:Abc"`
}

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
	OptName string `gorm:"column:OptName"`
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

type x_agent struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	UserId int `gorm:"column:UserId"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_agent_child struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	UserId int `gorm:"column:UserId"`
	ChildId int `gorm:"column:ChildId"`
	ChildLevel int `gorm:"column:ChildLevel"`
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
	ForClient int `gorm:"column:ForClient"`
	Memo string `gorm:"column:Memo"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_error struct {
	Id int `gorm:"column:Id"`
	FunName string `gorm:"column:FunName"`
	ErrCode int `gorm:"column:ErrCode"`
	ErrMsg string `gorm:"column:ErrMsg"`
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

type x_user struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	UserId int `gorm:"column:UserId"`
	State int `gorm:"column:State"`
	Account string `gorm:"column:Account"`
	Password string `gorm:"column:Password"`
	Token string `gorm:"column:Token"`
	NickName string `gorm:"column:NickName"`
	PhoneNum string `gorm:"column:PhoneNum"`
	Email string `gorm:"column:Email"`
	TopAgent int `gorm:"column:TopAgent"`
	Agents string `gorm:"column:Agents"`
	Agent int `gorm:"column:Agent"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_user_pool struct {
	UserId int `gorm:"column:UserId"`
	State int `gorm:"column:State"`
	Ip string `gorm:"column:Ip"`
	CreateTime string `gorm:"column:CreateTime"`
}

type x_user_score struct {
	Id int `gorm:"column:Id"`
	SellerId int `gorm:"column:SellerId"`
	ChannelId int `gorm:"column:ChannelId"`
	UserId int `gorm:"column:UserId"`
	Symbol string `gorm:"column:Symbol"`
	Amount float64 `gorm:"column:Amount"`
	FrozenAmount float64 `gorm:"column:FrozenAmount"`
	CreateTime string `gorm:"column:CreateTime"`
}

*/
