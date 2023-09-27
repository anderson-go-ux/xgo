package xgo

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
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
	conurl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", this.user, this.password, this.host, this.port, this.database)
	db, err := gorm.Open("mysql", conurl)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	db.DB().SetMaxIdleConns(this.connmaxidle)
	db.DB().SetMaxOpenConns(this.connmaxopen)
	db.DB().SetConnMaxIdleTime(time.Second * time.Duration(this.connmaxidletime))
	db.DB().SetConnMaxLifetime(time.Second * time.Duration(this.connmaxlifetime))
	this.db = db
	this.logmode = viper.GetBool(fmt.Sprint(cfgname, ".logmode"))
	db.LogMode(this.logmode)
	logs.Debug("连接数据库成功:", this.host, this.port, this.database)
}

func (this *XDb) conn() *sql.DB {
	return this.db.DB()
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
func (this *XDb) BeginTransaction() (*sql.Tx, error) {
	return this.db.DB().Begin()
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

	dbresult, err := this.db.DB().Query(sql, args...)
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
	data, err := this.db.DB().Exec(query, args...)
	if err != nil {
		logs.Error(query, args, err)
		return nil, err
	}
	return data, nil
}

// 执行查询 Query("select * from x_user where UserId = ?",12345)
func (this *XDb) Query(query string, args ...any) (*XMaps, error) {
	data, err := this.db.DB().Query(query, args...)
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
