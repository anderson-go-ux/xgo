package xgo

import (
	"database/sql"
	"encoding/json"
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

func (this *XDb) Init(cfgname string) {
	this.user = GetConfigString(fmt.Sprint(cfgname, ".user"), true, "")
	this.password = GetConfigString(fmt.Sprint(cfgname, ".password"), true, "")
	this.host = GetConfigString(fmt.Sprint(cfgname, ".host"), true, "")
	this.database = GetConfigString(fmt.Sprint(cfgname, ".database"), true, "")
	this.port = GetConfigInt(fmt.Sprint(cfgname, ".port"), true, 0)
	this.connmaxlifetime = GetConfigInt(fmt.Sprint(cfgname, ".connmaxlifetime"), true, 0)
	this.connmaxidletime = GetConfigInt(fmt.Sprint(cfgname, ".connmaxidletime"), true, 0)
	this.connmaxidle = GetConfigInt(fmt.Sprint(cfgname, ".connmaxidle"), true, 0)
	this.connmaxopen = GetConfigInt(fmt.Sprint(cfgname, ".connmaxopen"), true, 0)
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

func (this *XDb) Database() string {
	return this.database
}

func (this *XDb) Begin() (*sql.Tx, error) {
	return this.db.DB().Begin()
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

func (this *XDb) Exec(query string, args ...any) (sql.Result, error) {
	data, err := this.db.DB().Exec(query, args...)
	if err != nil {
		logs.Error(query, args, err)
		return nil, err
	}
	return data, nil
}

func (this *XDb) Query(query string, args ...any) (*XMaps, error) {
	data, err := this.db.DB().Query(query, args...)
	if err != nil {
		logs.Error(query, args, err)
		return nil, err
	}
	return this.GetResult(data), nil
}

type XMap struct {
	RawData map[string]any
}

type XMaps struct {
	RawData []XMap
}

func (this *XMap) FromBytes(bytes []byte) error {
	this.RawData = map[string]interface{}{}
	return json.Unmarshal(bytes, &this.RawData)
}

func (this *XMaps) FromBytes(bytes []byte) error {
	this.RawData = []XMap{}
	data := []map[string]interface{}{}
	json.Unmarshal(bytes, &data)
	for i := 0; i < len(data); i++ {
		XMap := XMap{RawData: data[i]}
		this.RawData = append(this.RawData, XMap)
	}
	return json.Unmarshal(bytes, &this.RawData)
}

func (this *XMap) FromObject(obj interface{}) error {
	bytes, err := json.Marshal(&obj)
	if err != nil {
		return err
	}
	this.RawData = map[string]interface{}{}
	json.Unmarshal(bytes, &this.RawData)
	return json.Unmarshal(bytes, &this.RawData)
}

func (this *XMaps) FromObjects(obj []interface{}) error {
	bytes, err := json.Marshal(&obj)
	if err != nil {
		return err
	}
	this.RawData = []XMap{}
	data := []map[string]interface{}{}
	json.Unmarshal(bytes, &data)
	for i := 0; i < len(data); i++ {
		XMap := XMap{RawData: data[i]}
		this.RawData = append(this.RawData, XMap)
	}
	return json.Unmarshal(bytes, &this.RawData)
}

func (this *XMaps) Maps() *[]map[string]interface{} {
	if this.RawData == nil {
		return nil
	}
	data := []map[string]interface{}{}
	for i := 0; i < len(this.RawData); i++ {
		data = append(data, *this.RawData[i].Map())
	}
	return &data
}

func (this *XMaps) Length() int {
	if this.RawData == nil {
		return 0
	}
	return len(this.RawData)
}

func (this *XMaps) Index(index int) *XMap {
	if this.RawData == nil {
		return nil
	}
	if index < 0 {
		return nil
	}
	if index >= len(this.RawData) {
		return nil
	}
	return &this.RawData[index]
}

func (this *XMaps) Remove(index int) {
	if this.RawData == nil {
		return
	}
	if index < 0 {
		return
	}
	if index >= len(this.RawData) {
		return
	}
	this.RawData = append(this.RawData[:index], this.RawData[index+1:]...)
}

func (this *XMaps) ForEach(cb func(*XMap) bool) {
	if this.RawData == nil {
		return
	}
	for i := 0; i < len(this.RawData); i++ {
		if !cb(&this.RawData[i]) {
			break
		}
	}
}

func (this *XMap) map_field(field string) interface{} {
	if this.RawData == nil {
		return nil
	}
	return (this.RawData)[field]
}

func (this *XMap) Map() *map[string]any {
	return &this.RawData
}

func (this *XMap) Int(field string) int {
	data := this.map_field(field)
	if data == nil {
		return 0
	}
	return int(ToInt(data))
}

func (this *XMap) Int32(field string) int32 {
	data := this.map_field(field)
	if data == nil {
		return 0
	}
	return int32(ToInt(data))
}

func (this *XMap) Int64(field string) int64 {
	data := this.map_field(field)
	if data == nil {
		return 0
	}
	return int64(ToInt(data))
}

func (this *XMap) Float32(field string) float32 {
	data := this.map_field(field)
	if data == nil {
		return 0
	}
	return float32(ToFloat(data))
}

func (this *XMap) Float64(field string) float64 {
	data := this.map_field(field)
	if data == nil {
		return 0
	}
	return ToFloat(data)
}

func (this *XMap) String(field string) string {
	data := this.map_field(field)
	if data == nil {
		return ""
	}
	return ToString(data)
}

func (this *XMap) Bytes(field string) []byte {
	data := this.map_field(field)
	if data == nil {
		return []byte{}
	}
	return []byte(ToString(data))
}

func (this *XMap) Delete(field string) {
	if this.RawData == nil {
		return
	}
	delete(this.RawData, field)
}

func (this *XMap) Set(field string, value interface{}) {
	if this.RawData == nil {
		return
	}
	this.RawData[field] = value
}

type XDbTable struct {
	db         *XDb
	tx         *sql.Tx
	tablename  []string
	wheregroup string
	groupopt   string
	wherestr   string
	wheredata  []interface{}
	selectstr  string
	orderby    string
	limit      int64
	offset     int64
	join       []string
}

func (this *XDb) Table(name string) *XDbTable {
	table := XDbTable{db: this, tablename: strings.Split(name, ",")}
	return &table
}

func (this *XDbTable) Tx(tx *sql.Tx) *XDbTable {
	this.tx = tx
	return this
}

func (this *XDbTable) Select(selectstr string) *XDbTable {
	this.selectstr = selectstr
	return this
}

func (this *XDbTable) OrderBy(selectstr string) *XDbTable {
	this.orderby = selectstr
	return this
}

func (this *XDbTable) Join(joinstr string) *XDbTable {
	this.join = append(this.join, joinstr)
	return this
}

func (this *XDbTable) Where(field interface{}, value ...interface{}) *XDbTable {
	if len(value) == 0 {
		if this.wherestr != "" {
			this.wherestr += " and "
		}
		if this.wheregroup != "" && this.groupopt == "" {
			this.groupopt += "and "
		}
		this.wherestr += fmt.Sprintf("(%v)", field)
	} else if len(value) == 1 {
		arrdata, ok := value[0].([]interface{})
		if ok && len(arrdata) == 0 {
			return this
		}
		if this.wherestr != "" {
			this.wherestr += " and "
		}
		if this.wheregroup != "" && this.groupopt == "" {
			this.groupopt += "and "
		}
		if !ok {
			this.wherestr += fmt.Sprintf("(%v)", field)
			this.wheredata = append(this.wheredata, value...)
		} else {
			v := "("
			for i := 0; i < len(arrdata); i++ {
				v += "?"
				if i < len(arrdata)-1 {
					v += ","
				}
			}
			v += ")"
			this.wherestr += fmt.Sprintf("(%v %v)", field, v)
			this.wheredata = append(this.wheredata, arrdata...)
		}
	} else {
		count := 0
		for _, char := range field.(string) {
			if char == '?' {
				count++
			}
		}
		if count > 1 {
			if this.wherestr != "" {
				this.wherestr += " and "
			}
			if this.wheregroup != "" && this.groupopt == "" {
				this.groupopt += "and "
			}
			this.wherestr += fmt.Sprintf("(%v)", field)
			this.wheredata = append(this.wheredata, value...)
		} else {
			if value[0] == value[1] {
				return this
			}
			if this.wherestr != "" {
				this.wherestr += " and "
			}
			if this.wheregroup != "" && this.groupopt == "" {
				this.groupopt += "and "
			}
			this.wherestr += fmt.Sprintf("(%v)", field)
			this.wheredata = append(this.wheredata, value[0])
		}
	}
	return this
}

func (this *XDbTable) Group() *XDbTable {
	if this.wherestr != "" {
		this.wheregroup += this.groupopt + fmt.Sprintf("(%v) ", this.wherestr)
		this.groupopt = ""
		this.wherestr = ""
	}
	fmt.Println(this.wheregroup + this.groupopt + this.wherestr)
	return this
}

func (this *XDbTable) Or(field interface{}, value interface{}, ignorevalue interface{}) *XDbTable {
	if value == ignorevalue {
		return this
	}
	if this.wherestr != "" {
		this.wherestr += " or "
	}
	if this.wheregroup != "" && this.groupopt == "" {
		this.groupopt += "or "
	}
	arrdata, ok := value.([]interface{})
	if !ok {
		this.wherestr += fmt.Sprintf("(%v)", field)
		this.wheredata = append(this.wheredata, value)
	} else {
		v := "("
		for i := 0; i < len(arrdata); i++ {
			v += "?"
			if i < len(arrdata)-1 {
				v += ","
			}
		}
		v += ")"
		this.wherestr += fmt.Sprintf("(%v %v)", field, v)
		this.wheredata = append(this.wheredata, arrdata...)
	}
	return this
}

func (this *XDbTable) First() (*XMap, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	sql := fmt.Sprintf("select %v from %v", this.selectstr, this.tablename[0])

	for i := 0; i < len(this.join); i++ {
		sql += fmt.Sprintf(" %v ", this.join[i])
	}

	if this.wherestr == "" {
		this.groupopt = ""
	}
	wherestr := this.wheregroup + this.groupopt + this.wherestr
	if wherestr != "" {
		sql += " where "
	}
	sql += wherestr
	if this.orderby != "" {
		sql += " order by "
		sql += this.orderby
	}
	sql += " limit 1"
	if this.tx != nil {
		data, err := this.tx.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		result := this.db.GetResult(data)
		if result.Length() == 0 {
			return nil, nil
		}
		return (*result).Index(0), nil
	} else {
		data, err := this.db.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		if data.Length() == 0 {
			return nil, nil
		}
		return (*data).Index(0), nil
	}

}

func (this *XDbTable) Limit(limit int64) *XDbTable {
	this.limit = limit
	return this
}

func (this *XDbTable) Offset(offset int64) *XDbTable {
	this.offset = offset
	return this
}

func (this *XDbTable) Find() (*XMaps, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	sql := ""
	for i := 0; i < len(this.tablename); i++ {
		sqlex := fmt.Sprintf("select %v from %v", this.selectstr, this.tablename[i])
		if len(this.tablename) == 1 {
			for i := 0; i < len(this.join); i++ {
				sqlex += fmt.Sprintf(" %v ", this.join[i])
			}
		}
		if this.wherestr == "" {
			this.groupopt = ""
		}
		wherestr := this.wheregroup + this.groupopt + this.wherestr
		if wherestr != "" {
			sqlex += " where "
		}
		sqlex += wherestr
		sql += sqlex
		if i < len(this.tablename)-1 {
			this.wheredata = append(this.wheredata, this.wheredata...)
			sql += " union "
		}
	}
	if this.orderby != "" {
		sql += " order by "
		sql += this.orderby
	}
	if this.limit > 0 {
		sql += fmt.Sprintf(" limit %v ", this.limit)
		if this.offset > 0 {
			sql += fmt.Sprintf("offset %v ", this.offset)
		}
	}
	if this.tx != nil {
		data, err := this.tx.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		return this.db.GetResult(data), nil
	} else {
		data, err := this.db.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

}

func (this *XDbTable) Insert(value interface{}) (int64, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	bytes, err := json.Marshal(&value)
	if err != nil {
		return 0, err
	}
	mapdata := map[string]interface{}{}
	err = json.Unmarshal(bytes, &mapdata)
	if err != nil {
		return 0, err
	}
	fields := ""
	placeholds := ""
	datas := []interface{}{}
	for k, v := range mapdata {
		fields += fmt.Sprintf("%v,", k)
		placeholds += "?,"
		datas = append(datas, v)
	}
	if len(datas) == 0 {
		return 0, nil
	}
	fields = fields[:len(fields)-1]
	placeholds = placeholds[:len(placeholds)-1]
	sql := fmt.Sprintf("insert into %v(%v)values(%v)", this.tablename[0], fields, placeholds)
	if this.tx != nil {
		resutl, err := this.tx.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		id, err := resutl.LastInsertId()
		return id, err
	} else {
		result, err := this.db.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return id, err
	}
}

func (this *XDbTable) Replace(value interface{}) (int64, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	bytes, err := json.Marshal(&value)
	if err != nil {
		return 0, err
	}
	mapdata := map[string]interface{}{}
	err = json.Unmarshal(bytes, &mapdata)
	if err != nil {
		return 0, err
	}
	fields := ""
	placeholds := ""
	datas := []interface{}{}
	for k, v := range mapdata {
		fields += fmt.Sprintf("%v,", k)
		placeholds += "?,"
		datas = append(datas, v)
	}
	if len(datas) == 0 {
		return 0, nil
	}
	fields = fields[:len(fields)-1]
	placeholds = placeholds[:len(placeholds)-1]
	sql := fmt.Sprintf("replace into %v(%v)values(%v)", this.tablename[0], fields, placeholds)
	if this.tx != nil {
		result, err := this.tx.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return id, err
	} else {
		result, err := this.db.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		return id, err
	}
}

func (this *XDbTable) Update(value interface{}) (int64, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	bytes, err := json.Marshal(&value)
	if err != nil {
		return 0, err
	}
	mapdata := map[string]interface{}{}
	err = json.Unmarshal(bytes, &mapdata)
	if err != nil {
		return 0, err
	}
	fields := ""
	datas := []interface{}{}
	for k, v := range mapdata {
		fields += fmt.Sprintf("%v = ?,", k)
		datas = append(datas, v)
	}
	if len(datas) == 0 {
		return 0, nil
	}
	fields = fields[:len(fields)-1]
	sql := fmt.Sprintf("update %v set %v ", this.tablename[0], fields)
	wherestr := this.wheregroup + this.groupopt + this.wherestr
	if wherestr != "" {
		sql += " where "
	}
	sql += wherestr
	datas = append(datas, this.wheredata...)
	if this.tx != nil {
		resutl, err := this.tx.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		cnt, err := resutl.RowsAffected()
		return cnt, err
	} else {
		resutl, err := this.db.Exec(sql, datas...)
		if err != nil {
			return 0, err
		}
		cnt, err := resutl.RowsAffected()
		return cnt, err
	}
}

func (this *XDbTable) Delete() (int64, error) {
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	sql := fmt.Sprintf("delete from %v ", this.tablename[0])
	wherestr := this.wheregroup + this.groupopt + this.wherestr
	if wherestr != "" {
		sql += " where "
	}
	sql += wherestr
	if this.tx != nil {
		result, err := this.tx.Exec(sql, this.wheredata...)
		if err != nil {
			return 0, err
		}
		cnt, err := result.RowsAffected()
		return cnt, err
	} else {
		result, err := this.db.Exec(sql, this.wheredata...)
		if err != nil {
			return 0, err
		}
		cnt, err := result.RowsAffected()
		return cnt, err
	}
}

func (this *XDbTable) Count() (int64, error) {
	str := this.selectstr
	this.selectstr = "count(*) as total"
	total, err := this.First()
	this.selectstr = str
	if err != nil {
		return 0, err
	}
	return total.Int64("total"), nil
}

func (this *XDbTable) PageData(page int, pagesize int) (*XMaps, error) {
	if page <= 0 {
		page = 1
	}
	if pagesize <= 0 {
		pagesize = 15
	}
	if this.selectstr == "" {
		this.selectstr = "*"
	}
	sql := ""
	for i := 0; i < len(this.tablename); i++ {
		sqlex := fmt.Sprintf("select %v from %v", this.selectstr, this.tablename[i])
		if len(this.tablename) == 1 {
			for i := 0; i < len(this.join); i++ {
				sqlex += fmt.Sprintf(" %v ", this.join[i])
			}
		}
		if this.wherestr == "" {
			this.groupopt = ""
		}
		wherestr := this.wheregroup + this.groupopt + this.wherestr
		if wherestr != "" {
			sqlex += " where "
		}
		sqlex += wherestr
		sql += sqlex
		if i < len(this.tablename)-1 {
			this.wheredata = append(this.wheredata, this.wheredata...)
			sql += " union "
		}
	}

	if this.orderby != "" {
		sql += " order by "
		sql += this.orderby
	}
	sql += fmt.Sprintf(" limit %v offset %v ", pagesize, (page-1)*pagesize)
	if this.tx != nil {
		data, err := this.tx.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		result := this.db.GetResult(data)
		if result.Length() == 0 {
			return result, nil
		}
		return result, nil
	} else {
		data, err := this.db.Query(sql, this.wheredata...)
		if err != nil {
			return nil, err
		}
		if data.Length() == 0 {
			return data, nil
		}
		return data, nil
	}
}

func (this *XDbTable) GetExportOptions() string {
	sql := fmt.Sprintf("SELECT COLUMN_NAME,COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS  WHERE TABLE_SCHEMA = '%s'  AND TABLE_NAME = '%s'", this.db.database, this.tablename[0])
	data, err := this.db.Query(sql)
	if err != nil {
		return ""
	}
	if data.Length() == 0 {
		return ""
	}
	obj := []map[string]interface{}{}
	data.ForEach(func(row *XMap) bool {
		field := row.String("COLUMN_NAME")
		name := row.String("COLUMN_COMMENT")
		values := map[string]interface{}{}
		if field == "SellerId" {
			name = "运营商"
			sellers, _ := this.db.Query("select SellerId,SellerName from x_seller")
			sellers.ForEach(func(row *XMap) bool {
				values[fmt.Sprint(row.String("SellerId"))] = row.String("SellerName")
				return true
			})
		}
		if field == "ChannelId" {
			name = "渠道"
			channels, _ := this.db.Query("select ChannelId,ChannelName from x_channel")
			channels.ForEach(func(row *XMap) bool {
				values[fmt.Sprint(row.String("ChannelId"))] = row.String("ChannelName")
				return true
			})
		}
		if field == "Id" {
			name = "Id"
		}
		if name == "" {
			name = field
		}
		obj = append(obj, H{
			"field":  field,
			"name":   name,
			"values": values,
		})
		return true
	})

	bytes, _ := json.MarshalIndent(&obj, "", "    ")
	return string(bytes)
}
