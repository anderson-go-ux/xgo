package xgo

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/logs"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type XDb struct {
	user            string
	password        string
	host            string
	port            int
	connmaxlifetime int
	database        string
	db              *gorm.DB
	connmaxidletime int
	connmaxidle     int
	connmaxopen     int
	logmode         bool
}

type XTx struct {
	Db *XDb
	Tx *sql.Tx
}

func (this *XTx) Table(table string) *XDbTable {
	t := this.Db.Table(table).Tx(this.Tx)
	return t
}

// 初始化db
func (this *XDb) Init(cfgname string) {
	this.user = GetConfigString(fmt.Sprint(cfgname, ".user"), true, "")
	this.password = GetConfigString(fmt.Sprint(cfgname, ".password"), true, "")
	this.host = GetConfigString(fmt.Sprint(cfgname, ".host"), true, "")
	this.database = GetConfigString(fmt.Sprint(cfgname, ".database"), true, "")
	this.port = int(GetConfigInt(fmt.Sprint(cfgname, ".port"), true, 0))
	this.connmaxlifetime = int(GetConfigInt(fmt.Sprint(cfgname, ".connmaxlifetime"), true, 0))
	this.connmaxidletime = int(GetConfigInt(fmt.Sprint(cfgname, ".connmaxidletime"), true, 0))
	this.connmaxidle = int(GetConfigInt(fmt.Sprint(cfgname, ".connmaxidle"), true, 0))
	this.connmaxopen = int(GetConfigInt(fmt.Sprint(cfgname, ".connmaxopen"), true, 0))
	conurl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.user, this.password, this.host, this.port, this.database)
	db, err := gorm.Open(mysql.Open(conurl), &gorm.Config{})
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	this.conn().SetMaxIdleConns(this.connmaxidle)
	this.conn().SetMaxOpenConns(this.connmaxopen)
	this.conn().SetConnMaxIdleTime(time.Second * time.Duration(this.connmaxidletime))
	this.conn().SetConnMaxLifetime(time.Second * time.Duration(this.connmaxlifetime))
	this.db = db
	this.logmode = viper.GetBool(fmt.Sprint(cfgname, ".logmode"))
	if this.logmode {
		this.db = this.db.Debug()
	}
	logs.Debug("connected to database successfully:", this.host, this.port, this.database)
}

func (this *XDb) conn() *sql.DB {
	sqlDB, err := this.db.DB()
	if err != nil {
		logs.Error("failed to get database connection:", err)
		return nil
	}
	return sqlDB
}

func (this *XDb) Gorm() *gorm.DB {
	return this.db
}

// 获取XDbTable
func (this *XDb) Table(name string) *XDbTable {
	table := XDbTable{db: this, tablename: strings.Split(name, ",")}
	return &table
}

// 获取连接的数据库
func (this *XDb) Database() string {
	return this.database
}

// 开始事务
func (this *XDb) Transaction(fc func(*XTx) error) {
	tx, err := this.conn().Begin()
	if err != nil {
		logs.Error("transaction error:", err)
		return
	}
	var fcerr error
	paniced := true
	defer func() {
		if paniced || fcerr != nil {
			tx.Rollback()
		}
	}()
	xtx := &XTx{Db: this, Tx: tx}
	fcerr = fc(xtx)
	if fcerr == nil {
		tx.Commit()
	}
	paniced = false
}

// 调用存储过程 eg: proc_test(int,int)  CallProcedure("proc_test",1,2)
func (this *XDb) CallProcedure(procname string, args ...interface{}) (*XMap, error) {
	sql := ""
	for i := 0; i < len(args); i++ {
		sql += "?,"
	}
	if len(sql) > 0 {
		sql = strings.TrimRight(sql, ",")
	}
	sql = fmt.Sprintf("call %s(%s)", procname, sql)

	dbresult, err := this.conn().Query(sql, args...)
	if err != nil {
		logs.Error(sql, args, err)
		return nil, err
	}
	if dbresult.Next() {
		data := make(map[string]interface{})
		fields, _ := dbresult.Columns()
		scans := make([]interface{}, len(fields))
		for i := range scans {
			scans[i] = &scans[i]
		}
		err := dbresult.Scan(scans...)
		if err != nil {
			return nil, err
		}
		ct, _ := dbresult.ColumnTypes()
		for i := range fields {
			if scans[i] != nil {
				typename := ct[i].DatabaseTypeName()
				if typename == "INT" || typename == "BIGINT" || typename == "TINYINT" || typename == "UNSIGNED BIGINT" || typename == "UNSIGNED" || typename == "UNSIGNED INT" {
					if reflect.TypeOf(scans[i]).Name() == "" {
						v, _ := strconv.ParseInt(string(scans[i].([]uint8)), 10, 64)
						data[fields[i]] = v
					} else {
						data[fields[i]] = scans[i]
					}
				} else if typename == "DOUBLE" || typename == "DECIMAL" {
					if reflect.TypeOf(scans[i]).Name() == "" {
						v, _ := strconv.ParseFloat(string(scans[i].([]uint8)), 64)
						data[fields[i]] = v
					} else {
						data[fields[i]] = scans[i]
					}
				} else {
					data[fields[i]] = string(scans[i].([]uint8))
				}
			} else {
				data[fields[i]] = nil
			}
		}
		dbresult.Close()
		XMap := XMap{RawData: data}
		return &XMap, nil
	}
	dbresult.Close()
	return nil, nil
}

// 获取sql.Rows返回的数据
func (this *XDb) GetResult(rows *sql.Rows) *XMaps {
	if rows == nil {
		return nil
	}
	data := []XMap{}
	for rows.Next() {
		data = append(data, XMap{RawData: *this.getone(rows)})
	}
	rows.Close()
	return &XMaps{RawData: data}
}

// 执行sql,无结果集返回 eg: Exec("update x_user set LoginIp = ?","127.0.0.1")
func (this *XDb) Exec(query string, args ...any) (sql.Result, error) {
	data, err := this.conn().Exec(query, args...)
	if err != nil {
		logs.Error(query, args, err)
		return nil, err
	}
	return data, nil
}

// 执行查询 Query("select * from x_user where UserId = ?",12345)
func (this *XDb) Query(query string, args ...any) (*XMaps, error) {
	data, err := this.conn().Query(query, args...)
	if err != nil {
		logs.Error(query, args, err)
		return nil, err
	}
	return this.GetResult(data), nil
}

func (this *XDb) getone(rows *sql.Rows) *map[string]any {
	data := make(map[string]interface{})
	fields, _ := rows.Columns()
	scans := make([]interface{}, len(fields))
	for i := range scans {
		scans[i] = &scans[i]
	}
	err := rows.Scan(scans...)
	if err != nil {
		logs.Error(err)
		return nil
	}
	ct, _ := rows.ColumnTypes()
	for i := range fields {
		if scans[i] != nil {
			typename := ct[i].DatabaseTypeName()
			if typename == "INT" || typename == "BIGINT" || typename == "TINYINT" || typename == "UNSIGNED BIGINT" || typename == "UNSIGNED" || typename == "UNSIGNED INT" {
				if reflect.TypeOf(scans[i]).Name() == "" {
					v, _ := strconv.ParseInt(string(scans[i].([]uint8)), 10, 64)
					data[fields[i]] = v
				} else {
					data[fields[i]] = scans[i]
				}
			} else if typename == "DOUBLE" || typename == "DECIMAL" {
				if reflect.TypeOf(scans[i]).Name() == "" {
					v, _ := strconv.ParseFloat(string(scans[i].([]uint8)), 64)
					data[fields[i]] = v
				} else {
					data[fields[i]] = scans[i]
				}
			} else if typename == "DATETIME" {
				timestr := string(scans[i].([]uint8))
				t, _ := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.Local)
				r := t.UTC().Format("2006-01-02T15:04:05Z")
				data[fields[i]] = r
			} else if typename == "DATE" {
				timestr := string(scans[i].([]uint8))
				timestr += " 00:00:00"
				t, _ := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.Local)
				r := t.UTC().Format("2006-01-02T15:04:05Z")
				data[fields[i]] = r
			} else {
				data[fields[i]] = string(scans[i].([]uint8))
			}
		} else {
			data[fields[i]] = nil
		}
	}
	return &data
}

func (this *XDb) Union(tables ...interface{}) (*XMaps, error) {
	sql := ""
	data := []interface{}{}
	for i := 0; i < len(tables); i++ {
		t := tables[i].(*XDbTable)
		if sql != "" {
			sql += " union "
		}
		s, d := t.GetQuery()
		sql += s
		for j := 0; j < len(d); j++ {
			data = append(data, d[j])
		}
	}
	rows, err := this.conn().Query(sql, data...)
	if err != nil {
		logs.Error(err, "|", sql, data)
		return nil, err
	}
	return this.GetResult(rows), nil
}
