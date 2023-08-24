package xgo

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	crand "crypto/rand"
	mrand "math/rand"

	"github.com/beego/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

type H map[string]any

func Init() {
	mrand.NewSource(time.Now().UnixNano())
	gin.SetMode(gin.ReleaseMode)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(5)
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
	env = GetConfigString("server.env", true, "")
	project = GetConfigString("server.project", true, "")
	ipdata = GetConfigString("server.ipdata", false, "")
}

func Env() string {
	return env
}

func Prjoect() string {
	return project
}

func Run() {
	for {
		time.Sleep(time.Second * 1)
	}
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func InterfaceToString(v interface{}) string {
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
	}
	return ""
}

func GetMapString(mp *map[string]interface{}, field string) string {
	if mp == nil {
		return ""
	}
	v := (*mp)[field]
	return InterfaceToString(v)
}

func InterfaceToInt(v interface{}) int64 {
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
	}
	return 0
}

func GetMapInt(mp *map[string]interface{}, field string) int64 {
	if mp == nil {
		return 0
	}
	v := (*mp)[field]
	return InterfaceToInt(v)
}

func InterfaceToFloat(v interface{}) float64 {
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
	}
	return 0
}

func GetMapFloat(mp *map[string]interface{}, field string) float64 {
	if mp == nil {
		return 0
	}
	v := (*mp)[field]
	return InterfaceToFloat(v)
}

func onetimepassword(key []byte, value []byte) uint32 {
	hmacSha1 := hmac.New(sha1.New, key)
	hmacSha1.Write(value)
	hash := hmacSha1.Sum(nil)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := touint32(hashParts)
	pwd := number % 1000000
	return pwd
}

func touint32(bytes []byte) uint32 {
	return (uint32(bytes[0]) << 24) + (uint32(bytes[1]) << 16) +
		(uint32(bytes[2]) << 8) + uint32(bytes[3])
}

func tobytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func GetGoogleCode(secret string) int32 {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		logs.Error(err)
		return 0
	}
	epochSeconds := time.Now().Unix() + 0
	return int32(onetimepassword(key, tobytes(epochSeconds/30)))
}

func VerifyGoogleCode(secret string, code string) bool {
	nowcode := GetGoogleCode(secret)
	if fmt.Sprint(nowcode) == code {
		return true
	}
	return false
}

func googlerandstr(strSize int) string {
	dictionary := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var bytes = make([]byte, strSize)
	_, _ = crand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func NewGoogleSecret() string {
	return strings.ToUpper(googlerandstr(32))
}

func ReadAllText(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return string(bytes)
}

func TimeStampToLocalDate(tvalue int64) string {
	if tvalue == 0 {
		return ""
	}
	tm := time.Unix(tvalue, 0)
	tstr := tm.Format("2006-01-02")
	return strings.Split(tstr, " ")[0]
}

func LocalDateToTimeStamp(timestr string) int64 {
	t, _ := time.ParseInLocation("2006-01-02", timestr, time.Local)
	return t.Local().Unix()
}

func TimeStampToLocalTime(tvalue int64) string {
	if tvalue == 0 {
		return ""
	}
	tm := time.Unix(tvalue, 0)
	tstr := tm.Format("2006-01-02 15:04:05")
	return tstr
}

func LocalTimeToTimeStamp(timestr string) int64 {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.Local)
	return t.Local().Unix()
}

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

func GetLocalTime() string {
	tm := time.Now()
	return tm.In(time.Local).Format("2006-01-02 15:04:05")
}

func GetLocalDate() string {
	tm := time.Now()
	return tm.In(time.Local).Format("2006-01-02")
}

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

type DBError struct {
	ErrCode int
	ErrMsg  string
}

func GetDbError(data *map[string]interface{}) *DBError {
	err := DBError{}
	code, codeok := (*data)["errcode"]
	if !codeok {
		return nil
	}
	err.ErrCode = int(InterfaceToInt(code))
	if err.ErrCode == 0 {
		return nil
	}
	msg, msgok := (*data)["errmsg"]
	if msgok {
		err.ErrMsg = InterfaceToString(msg)
	}
	return &err
}

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

func BackupDb(db *XDb, path string) {
	if env != "dev" {
		return
	}
	var strall string
	var tables []string
	data, _ := db.Query("SHOW FULL TABLES")
	data.ForEach(func(xd *XDbData) bool {
		for k, v := range *xd.RawData() {
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
				s += ";\r\n\r\n"
				strall += s

			}
		}
		return true
	})
	strall += "/*\r\n"
	for i := 0; i < len(tables); i++ {
		td, _ := db.Query(fmt.Sprintf("DESCRIBE %v", tables[i]))
		strall += fmt.Sprintf("type %v struct {\r\n", tables[i])
		td.ForEach(func(xd *XDbData) bool {
			sname := xd.String("Type")
			tname := ""
			if strings.Index(sname, "int") == 0 || strings.Index(sname, "bigint") == 0 || strings.Index(sname, "unsigned") == 0 || strings.Index(sname, "timestamp") == 0 {
				tname = "int"
			} else if strings.Index(sname, "varchar") == 0 || strings.Index(sname, "datetime") == 0 || strings.Index(sname, "date") == 0 || strings.Index(sname, "text") == 0 {
				tname = "string"
			} else if strings.Index(sname, "decimal") == 0 {
				tname = "float64"
			}
			strall += fmt.Sprintf("\t%v %v `gorm:\"column:%v\"`\r\n", xd.String("Field"), tname, xd.String("Field"))
			return true
		})
		strall += "}\r\n\r\n"
	}
	strall += "*/\r\n"
	file, _ := os.OpenFile(path, os.O_TRUNC|os.O_CREATE, 0666)
	write := bufio.NewWriter(file)
	write.WriteString(strall)
	write.Flush()
}

func pcks5padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func DesCbcEncrypt(data, key []byte, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	data = pcks5padding(data, bs)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(data))
	blockMode.CryptBlocks(out, data)
	return out, nil
}

func Base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}
