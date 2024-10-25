package xgo

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

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

// 设置事务连接,如果设置了事务连接,本对象的所有操作都在事务连接上执行,一定要commit或者rollback事务,否则连接无法释放
// tx参数通过 XDb BeginTransaction 方法获得
func (this *XDbTable) Tx(tx *sql.Tx) *XDbTable {
	this.tx = tx
	return this
}

// 选择项,默认Select("*")  eg:Select("count(UserId) as UserCount")
func (this *XDbTable) Select(selectstr string) *XDbTable {
	if selectstr == "" {
		this.selectstr = selectstr
	} else {
		if this.selectstr != "" {
			this.selectstr += ","
		}
		this.selectstr += selectstr
	}
	return this
}

// 结果排序 eg: OrderBy("Id desc")  OrderBy("Amount desc,Id desc")
func (this *XDbTable) OrderBy(selectstr string) *XDbTable {
	this.orderby = selectstr
	return this
}

// join表 eg: Join("left join x_user on x_user.UserId = x_table.UserId")
func (this *XDbTable) Join(joinstr string) *XDbTable {
	this.join = append(this.join, joinstr)
	return this
}

// where条件,条件连接符为and  Where("UserId = 123").Where("Account = 'abc'")  UserId = 123 and Account = 'abc'
/*
	eg:
		Where("UserId = 123 and Amount < 100")
		where("UserId = ? and Amount < ?",123,100)
		Where("Amount < ?",xxx,50) //当field只有一个? ,value有两个值,如果两个值相等,则此忽略此条件 ,当 xxx等于50时,忽略本句
		Where("Memo like ?","%abc%")
		Where("Memo in ?",[]int{1,2,3})
*/
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

// 将已添加的条件打包成一个整体条件 where("UserId = 123").Or("Account = 'abc'").Group().where("test1 = 1").or("test2=3")
// (UserId = 123 or Account = 'abc') and (test1 = 1 or test2=3)
func (this *XDbTable) Group() *XDbTable {
	if this.wherestr != "" {
		this.wheregroup += this.groupopt + fmt.Sprintf("(%v) ", this.wherestr)
		this.groupopt = ""
		this.wherestr = ""
	}
	return this
}

// 同where,连接符是or Or("UserId = 123").Or("Account = 'abc'")  UserId = 123 or Account = 'abc'
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

// 找第一个,只找一个
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

// 限制返回数量
func (this *XDbTable) Limit(limit int64) *XDbTable {
	this.limit = limit
	return this
}

// 跳过行数
func (this *XDbTable) Offset(offset int64) *XDbTable {
	this.offset = offset
	return this
}

// 查找符合条件的所有数据
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
func (this *XDbTable) GetQuery() (string, []interface{}) {
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
	return sql, this.wheredata
}

// 插入数据
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

// replace数据
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

// 更新数据
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

// 删除数据
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

// 获取符合条件的数量
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

// 获取分页数据
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

// 获取表的导出配置
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
