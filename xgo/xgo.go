package xgo

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"math/rand"
	mrand "math/rand"

	"github.com/beego/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"github.com/yinheli/qqwry"
)

var env string
var project string
var ipdata string

/*
go get github.com/beego/beego/logs
go get github.com/spf13/viper
go get github.com/gin-gonic/gin
go get github.com/go-redis/redis
go get github.com/garyburd/redigo/redis
go get github.com/go-sql-driver/mysql
go get github.com/satori/go.uuid
go get github.com/gorilla/websocket
go get github.com/jinzhu/gorm
go get github.com/imroc/req
go get github.com/go-playground/validator/v10
go get github.com/go-playground/universal-translator
go get code.google.com/p/mahonia
go get github.com/360EntSecGroup-Skylar/excelize
go clean -modcache
*/
var TimeLayout string = "2006-01-02 15:04:05"
var DateLayout string = "2006-01-02"

// 通用map定义
type H map[string]any

// 初始化
func Init() {
	mrand.NewSource(time.Now().UnixNano())
	gin.SetMode(gin.ReleaseMode)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogger(logs.AdapterFile, `{"filename":"_log/logfile.log","maxsize":10485760}`)
	logs.SetLogger(logs.AdapterConsole, `{"color":true}`)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		logs.Error("读取配置文件失败", err)
		return
	}
	snowflakenode := GetConfigInt("server.snowflakenode", true, 0)
	if snowflakenode != 0 {
		newIdWorker(snowflakenode)
	}

	env = GetConfigString("server.env", true, "")
	project = GetConfigString("server.project", true, "")
	ipdata = GetConfigString("server.ipdata", false, "")
	rpcport := viper.GetInt("server.rpc.port")
	if rpcport > 0 {
		go func() {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcport))
			if err != nil {
				logs.Error("RPC 开启失败:", err)
				return
			}
			go func() {
				for {
					conn, err := listener.Accept()
					if err != nil {
						continue
					}
					go rpc.ServeConn(conn)
				}
			}()
		}()
	}
}

// 获取配置文件配置的env
func Env() string {
	return env
}

// 获取配置文件配置的project
func Prjoect() string {
	return project
}

// 阻塞,防止主线程退出
func Run() {
	for {
		time.Sleep(time.Second * 1)
	}
}

// 获取字符串md5值 eg:test -> 098f6bcd4621d373cade4e832627b4f6
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// interface转string
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return fmt.Sprint(v.(int))
	case int32:
		return fmt.Sprint(v.(int32))
	case int64:
		return fmt.Sprint(v.(int64))
	case float32:
		return fmt.Sprint(v.(float32))
	case float64:
		return fmt.Sprint(v.(float64))
	default:
		if bytes, ok := v.([]byte); ok {
			return string(bytes)
		}
	}
	return ""
}

// interface转int64
func ToInt(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch v.(type) {
	case string:
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return 0
		}
		return i
	case int:
		return int64(v.(int))
	case int32:
		return int64(v.(int32))
	case int64:
		return int64(v.(int64))
	case float32:
		return int64(v.(float32))
	case float64:
		return int64(v.(float64))
	default:
		if bytes, ok := v.([]byte); ok {
			i, err := strconv.ParseInt(string(bytes), 10, 64)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

// interface转float64
func ToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch v.(type) {
	case string:
		i, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0
		}
		return i
	case int:
		return float64(v.(int))
	case int32:
		return float64(v.(int32))
	case int64:
		return float64(v.(int64))
	case float32:
		return float64(v.(float32))
	case float64:
		return v.(float64)
	default:
		if bytes, ok := v.([]byte); ok {
			i, err := strconv.ParseFloat(string(bytes), 64)
			if err != nil {
				return 0
			}
			return i
		}
	}
	return 0
}

// 验证谷歌验证码
func VerifyGoogleCode(secret string, code string) bool {
	return totp.Validate(code, secret)
}

// 创建新的google secret
func NewGoogleSecret(Issuer string, AccountName string) (string, string) {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: AccountName,
	})
	return key.Secret(), key.URL()
}

func GetGoogleQrCodeUrl(secret string, issuer string, accountname string) string {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountname,
		Secret:[]byte(secret),
	})
	return key.URL()
}

// 读取文件全部文本
func ReadAllText(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return string(bytes)
}

// 时间戳转本地日期 eg: 1609459200 -> 2021-01-01
func TimeStampToLocalDate(tvalue int64) string {
	if tvalue == 0 {
		return ""
	}
	tm := time.Unix(tvalue, 0)
	tstr := tm.Format("2006-01-02")
	return strings.Split(tstr, " ")[0]
}

// 本地日期转时间戳(秒) eg: 2021-01-01 -> 1609459200
func LocalDateToTimeStamp(timestr string) int64 {
	t, _ := time.ParseInLocation("2006-01-02", timestr, time.Local)
	return t.Local().Unix()
}

// 时间戳(秒)转本地时间 eg: 1609459200 -> 2021-01-01 08:00:00
func TimeStampToLocalTime(tvalue int64) string {
	if tvalue == 0 {
		return ""
	}
	tm := time.Unix(tvalue, 0)
	tstr := tm.Format("2006-01-02 15:04:05")
	return tstr
}

// 本地时间转时间戳(秒) eg: 2021-01-01 08:00:00 -> 1609459200
func LocalTimeToTimeStamp(timestr string) int64 {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.Local)
	return t.Local().Unix()
}

// 本地时间转utc时间 eg: 2021-01-01 08:00:00 -> 2021-01-01T00:00:00Z
func LocalTimeToUtc(timestr string) string {
	if len(timestr) == 0 {
		return timestr
	}
	if len(timestr) == 10 {
		timestr = timestr + " 00:00:00"
	}
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.Local)
	r := t.UTC().Format("2006-01-02T15:04:05Z")
	return r
}

// utc时间转本地时间 eg: 2021-01-01T00:00:00Z -> 2021-01-01 08:00:00
func UtcToLocalTime(timestr string) string {
	if len(timestr) == 0 {
		return ""
	}
	t, err := time.Parse(time.RFC3339, timestr)
	if err != nil {
		return ""
	}
	localTime := t.Local()
	return localTime.In(time.Local).Format("2006-01-02 15:04:05")
}

// 获取本地时间 eg: 2021-01-01 00:00:00
func GetLocalTime() string {
	tm := time.Now()
	return tm.In(time.Local).Format("2006-01-02 15:04:05")
}

// 获取本地日期 eg: 2021-01-01
func GetLocalDate() string {
	tm := time.Now()
	return tm.In(time.Local).Format("2006-01-02")
}

// go对象转map
func ObjectToMap(obj any) *map[string]interface{} {
	bytes, err := json.Marshal(obj)
	if err != nil {
		logs.Error("ObjectToMap:", err)
		return nil
	}
	data := map[string]interface{}{}
	json.Unmarshal(bytes, &data)
	return &data
}

// 获取ip地理位置
func GetIpLocation(ip string) string {
	if ipdata == "" {
		return ""
	}
	iptool := qqwry.NewQQwry("./config/ipdata.dat")
	if strings.Index(ip, ".") > 0 {
		iptool.Find(ip)
		return fmt.Sprintf("%s %s", iptool.Country, iptool.City)
	} else {
		return ""
	}
}

// 备份数据库
// db: 数据库连接
// path: 备份文件路径
func BackupDb(db *XDb, path string) {
	if env != "dev" {
		return
	}
	var strall string
	var tables []string
	tabledata, _ := db.Query("SHOW FULL TABLES")
	tabledata.ForEach(func(xd *XMap) bool {
		for k, v := range *xd.Map() {
			if strings.Index(k, "Tables_in_") >= 0 {
				tables = append(tables, v.(string))
				td, _ := db.Query(fmt.Sprint("show create table ", v))
				s := td.Index(0).String("Create Table")
				s = strings.Replace(s, "`", "", -1)
				s = strings.Replace(s, "DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci", "", -1)
				s = strings.Replace(s, "CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci", "", -1)
				s = strings.Replace(s, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS ", -1)

				s = strings.Replace(s, "DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci", "", -1)
				s = strings.Replace(s, "COLLATE utf8mb4_general_ci", "", -1)
				s = strings.Replace(s, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS ", -1)
				s = strings.Replace(s, "IF NOT EXISTS  IF NOT EXISTS", "IF NOT EXISTS", -1)

				idx := strings.Index(s, "AUTO_INCREMENT=")
				if idx > 0 {
					es := s[idx:]
					eidx := strings.Index(es, " ")
					es = es[:eidx]
					s = strings.Replace(s, es, "AUTO_INCREMENT=0 ", -1)
				}
				s = strings.Replace(s, "  ROW_", " ROW_", -1)
				s = strings.Replace(s, "  ROW_", " ROW_", -1)
				s += ";\r\n\r\n"
				strall += s

			}
		}
		return true
	})
	procdata, _ := db.Query(`SHOW PROCEDURE STATUS LIKE "%x_%"`)
	procdata.ForEach(func(xd *XMap) bool {
		if xd.String("Db") == db.database {
			strall += fmt.Sprintf("DROP PROCEDURE IF EXISTS `%s`;\r\ndelimiter ;;\r\n", xd.String("Name"))
			createdata, _ := db.Query(fmt.Sprintf("SHOW CREATE PROCEDURE %v", xd.String("Name")))
			createsql := createdata.Index(0).String("Create Procedure")
			re := regexp.MustCompile(" DEFINER=`[^`]+`@`[^`]+`")
			matches := re.FindString(createsql)
			if matches != "" {
				createsql = strings.Replace(createsql, matches, "", -1)
			}
			strall += createsql
			strall += "\r\n;;\r\ndelimiter ;\r\n\r\n"
		}
		return true
	})

	// strall += "/*\r\n"
	// for i := 0; i < len(tables); i++ {
	// 	td, _ := db.Query(fmt.Sprintf("DESCRIBE %v", tables[i]))
	// 	strall += fmt.Sprintf("type %v struct {\r\n", tables[i])
	// 	td.ForEach(func(xd *XMap) bool {
	// 		sname := xd.String("Type")
	// 		tname := ""
	// 		if strings.Index(sname, "int") == 0 || strings.Index(sname, "bigint") == 0 || strings.Index(sname, "unsigned") == 0 || strings.Index(sname, "timestamp") == 0 {
	// 			tname = "int"
	// 		} else if strings.Index(sname, "varchar") == 0 || strings.Index(sname, "datetime") == 0 || strings.Index(sname, "date") == 0 || strings.Index(sname, "text") == 0 {
	// 			tname = "string"
	// 		} else if strings.Index(sname, "decimal") == 0 {
	// 			tname = "float64"
	// 		}
	// 		strall += fmt.Sprintf("\t%v %v `gorm:\"column:%v\"`\r\n", xd.String("Field"), tname, xd.String("Field"))
	// 		return true
	// 	})
	// 	strall += "}\r\n\r\n"
	// }
	// strall += "*/\r\n"
	file, _ := os.OpenFile(path, os.O_TRUNC|os.O_CREATE, 0666)
	write := bufio.NewWriter(file)
	write.WriteString(strall)
	write.Flush()
}

// des cbc解密
// data:密文
// key:密钥
// iv:偏移量
func DesCbcEncrypt(data []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	padding := bs - len(data)%bs
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	data = append(data, padText...)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(data))
	blockMode.CryptBlocks(out, data)
	return out, nil
}

// base64字符串解码获得[]byte
func Base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}

// 判断字符串是否包含小写字母
func StrContainsLower(str string) bool {
	for _, char := range str {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

// 判断字符串是否包含大写字母
func StrContainsUpper(str string) bool {
	for _, char := range str {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

// 判断支付是否包含数字
func StrContainsDigit(str string) bool {
	for _, char := range str {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

//导出excel
//filename:excel文件名
//edata:要导出数据
//options:导出配置
/*
	options格式:
	[
		{
			"field":"State", edata中的字段名
			"name":"状态", 导出到excel的列名
			"values":{"1":"男","2":"女"}} edata中的值转换成excel中的值 eg:edata中的State字段值为1,则导出到excel中的值为男
			"time":1, 可选项,如果设置了time,则会把时间转换成本地时间 eg:2021-01-01T00:00:00Z -> 2021-01-01 08:00:00
			"date":1, 可选项,如果设置了date,则会把时间转换成本地日期 eg:2021-01-01T00:00:00Z -> 2021-01-01
		}
	]
*/
func Export(filename string, edata *XMaps, options string) string {
	excel := excelize.NewFile()
	jopt := []map[string]interface{}{}
	json.Unmarshal([]byte(options), &jopt)
	columns := []string{}
	for i := 0; i < len(jopt); i++ {
		columns = append(columns, jopt[i]["name"].(string))
	}
	excel.SetSheetRow("Sheet1", "A1", &columns)
	drow := 0
	edata.ForEach(func(row *XMap) bool {
		rowdata := []string{}
		for i := 0; i < len(jopt); i++ {
			bytes, _ := json.Marshal(jopt[i]["values"])
			desc := map[string]interface{}{}
			json.Unmarshal(bytes, &desc)
			v := row.String(jopt[i]["field"].(string))
			if jopt[i]["time"] != nil {
				v = UtcToLocalTime(v)
			}
			if jopt[i]["date"] != nil {
				v = UtcToLocalTime(v)
				tm := LocalTimeToTimeStamp(v)
				v = TimeStampToLocalDate(tm)
			}
			fv := desc[v]
			if fv == nil {
				rowdata = append(rowdata, v)
			} else {
				rowdata = append(rowdata, ToString(fv.(string)))
			}
		}
		excel.SetSheetRow("Sheet1", fmt.Sprintf("A%d", drow+2), &rowdata)
		drow++
		return true
	})
	filename = filename + ".xlsx"
	excel.SaveAs(filename)
	return filename
}

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(bytes)
}

func GetId() int64 {
	return snow_worker.GetId()
}
