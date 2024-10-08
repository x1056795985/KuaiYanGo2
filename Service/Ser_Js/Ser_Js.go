package Ser_Js

// https://blog.csdn.net/wyongqing/article/details/124704136   参考地址
import (
	. "EFunc/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"regexp"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/new/app/logic/common/mqttClient"
	"server/new/app/logic/common/publicData"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

func JS引擎初始化_用户(AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken, 局_PublicJs *DB.DB_PublicJs) *goja.Runtime {
	vm := goja.New() // 创建engine实例
	_ = vm.Set("$用户在线信息", 在线信息)

	局_AppInfo := vm.NewObject()
	局_AppInfo.Set("AppId", AppInfo.AppId)
	局_AppInfo.Set("AppName", AppInfo.AppName)
	局_AppInfo.Set("Status", AppInfo.Status)
	局_AppInfo.Set("VipData", AppInfo.VipData)
	_ = vm.Set("$应用信息", 局_AppInfo)

	console := vm.NewObject()
	_ = console.Set("log", jS_log)
	_ = vm.Set("console", console)

	_ = vm.Set("$程序_延时", jS_程序_延时)
	_ = vm.Set("$api_用户Id取详情", jS_用户Id取详情)
	_ = vm.Set("$api_卡号Id取详情", jS_卡号Id取详情)
	_ = vm.Set("$api_取软件用户详情", jS_取软件用户详情)

	_ = vm.Set("$api_用户Id增减余额", jS_用户Id增减余额)
	_ = vm.Set("$api_用户Id增减积分", jS_用户Id增减积分)
	_ = vm.Set("$api_用户Id增减时间点数", jS_用户Id增减时间点数)
	_ = vm.Set("$api_读公共变量", jS_读公共变量)
	_ = vm.Set("$api_置公共变量", jS_置公共变量)
	_ = vm.Set("$api_网页访问_GET", jS_网页访问_GET)
	_ = vm.Set("$api_网页访问_POST", jS_网页访问_POST)
	_ = vm.Set("$api_置动态标记", jS_置在线动态标签)
	_ = vm.Set("$api_执行SQL查询", jS_执行SQL查询)
	_ = vm.Set("$api_执行SQL功能", jS_执行SQL功能)
	_ = vm.Set("$api_任务池_任务创建", jS_任务池_任务创建)
	_ = vm.Set("$api_任务池_任务查询", jS_任务池_任务查询)
	_ = vm.Set("$api_短信发送", jS_任务池_任务查询)
	_ = vm.Set("$api_用户名或卡号取uid", jS_用户名或卡号取uid)
	_ = vm.Set("$api_取用户云配置", jS_取用户云配置)
	_ = vm.Set("$api_置用户云配置", jS_置用户云配置)
	_ = vm.Set("$api_取缓存", jS_取缓存)
	_ = vm.Set("$api_置缓存", jS_置缓存)
	_ = vm.Set("$api_置黑名单", jS_置黑名单)
	_ = vm.Set("$api_mqtt发送消息", jS_mqtt发送消息)
	_ = vm.Set("$api_任务池Uuid添加到队列", jS_任务池Uuid添加到队列)

	_ = vm.Set("$api_编码_BASE64编码", B编码_BASE64编码)
	_ = vm.Set("$api_编码_BASE64解码", B编码_BASE64解码)
	_ = vm.Set("$api_字节集_十六进制到字节集", Z字节集_十六进制到字节集)
	_ = vm.Set("$api_字节集_字节集到十六进制", Z字节集_字节集到十六进制)
	_ = vm.Set("$api_文本_取文本右边", W文本_取文本右边)
	_ = vm.Set("$api_文本_取文本左边", W文本_取文本左边)
	_ = vm.Set("$api_文本_取出中间文本", W文本_取出中间文本)
	_ = vm.Set("$api_生成二维码并转base64", rmbPay.L_rmbPay.S生成二维码并转base64)

	//处理载入外部js文件  'import "@/utils/utils";
	if strings.Index(局_PublicJs.Value, "import '") != -1 || strings.Index(局_PublicJs.Value, `import "`) != -1 {
		// 导入外部的模块
		re := regexp.MustCompile(`\n@?import\s+['|"](.*?)['|"]`)
		result := re.FindAllStringSubmatch(`\n`+局_PublicJs.Value, -1)
		for i, _ := range result {
			局_临时文本 := result[i][1]
			if W文本_取左边(局_临时文本, 4) == `http` {
				//https://lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js
				局_本地路径 := global.GVA_CONFIG.Q取运行目录 + "/云函数/lib/" + W文本_取文本右边(局_临时文本, "//")
				if !W文件_是否存在(局_本地路径) || W文本_取左边(result[i][0], 1) == "@" {
					局_js, err2 := req.C().EnableInsecureSkipVerify().R().Get(局_临时文本)
					if err2 == nil {
						_ = M目录_创建(W文件_取父目录(局_本地路径))
						_ = W文件_写到文件(局_本地路径, 局_js.Bytes())
					}
				}
				局_PublicJs.Value = strings.Replace(局_PublicJs.Value, strings.TrimSpace(result[i][0]), W文件_读入文本(局_本地路径), 1)
			} else {
				//lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js
				局_本地路径 := global.GVA_CONFIG.Q取运行目录 + "/云函数/" + 局_临时文本
				局_PublicJs.Value = strings.Replace(局_PublicJs.Value, strings.TrimSpace(result[i][0]), W文件_读入文本(局_本地路径), 1)
			}
		}
	}

	return vm
}
func JS引擎初始化_任务池Hook处理(AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken, Hook函数, 任务数据 string, 局_任务状态 int) (string, int, error) {
	局_PublicJs, err := Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_任务池Hook函数, Hook函数)
	if err != nil {
		return "", 局_任务状态, err
	}

	vm := JS引擎初始化_用户(AppInfo, 在线信息, &局_PublicJs)
	_ = vm.Set("$拦截原因", "")
	_ = vm.Set("$任务状态", 局_任务状态)
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok2 := err.(*goja.Exception); ok2 {
		return "", 局_任务状态, errors.New("JS代码运行失败:" + 局_详细错误.String())
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		return "", 局_任务状态, errors.New("Js中没有[" + 局_PublicJs.Name + "()]函数")
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		return "", 局_任务状态, errors.New("js绑定函数到变量失败")
	}

	局_return := 局_待执行js函数名(任务数据).(string)
	局_拦截原因 := vm.Get("$拦截原因").Export().(string)
	局_任务状态64, ok := vm.Get("$任务状态").Export().(int64) //goja js整数到go整数转换必须是int64 否则恐慌报错
	if ok {
		局_任务状态 = int(局_任务状态64)
	}
	if 局_拦截原因 != "" {
		return "", 局_任务状态, errors.New(局_拦截原因)
	}

	return 局_return, 局_任务状态, nil

}
func JS引擎初始化_ApiHook处理(AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken, Hook函数 string, 明文信息 string, c *gin.Context) (局_明文信息 string, err error) {
	defer func() {
		err2 := recover() // recover()内置函数，可以捕获到异常
		if err2 != nil {  //说明捕获到错误
			err = errors.New("js函数错误:" + fmt.Sprintln(err2))
		}
	}()
	局_明文信息 = 明文信息
	局_PublicJs, err := Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_ApiHook函数, Hook函数)
	if err != nil {
		return
	}

	vm := JS引擎初始化_用户(AppInfo, 在线信息, &局_PublicJs)
	_ = vm.Set("$拦截原因", "")
	headers := make([]string, 0, len(c.Request.Header))
	for key, values := range c.Request.Header {
		header := key + ": " + strings.Join(values, ", ")
		headers = append(headers, header)
	}
	Request := map[string]interface{}{"Url": c.Request.URL, "Header": headers, "Host": c.Request.Host, "Body": 明文信息, "Method": c.Request.Method}
	_ = vm.Set("$Request", Request)

	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok2 := err.(*goja.Exception); ok2 {
		err = errors.New("JS代码运行失败:" + 局_详细错误.String())
		return
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		err = errors.New("Js中没有[" + 局_PublicJs.Name + "()]函数")
		return
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		err = errors.New("js绑定函数到变量失败")
		return

	}

	局_明文信息 = 局_待执行js函数名(明文信息).(string)
	局_拦截原因 := vm.Get("$拦截原因").Export().(string)

	if 局_拦截原因 != "" {
		err = errors.New(局_拦截原因)
		return
	}
	return 局_明文信息, nil

}

func jS_log(call goja.FunctionCall) goja.Value {
	str := call.Argument(0)
	fmt.Print(str.String())
	global.GVA_LOG.Info(str.String())
	return str
}

func jS_用户Id取详情(局_在线信息 DB.DB_LinksToken) DB.DB_User {
	var 局_用户详情 DB.DB_User
	局_用户详情, ok := Ser_User.Id取详情(局_在线信息.Uid)
	if ok {
		return 局_用户详情
	}
	return 局_用户详情
}
func jS_程序_延时(毫秒数 int64) bool {
	time.Sleep(time.Duration(毫秒数) * time.Millisecond)
	return true
}
func jS_卡号Id取详情(局_在线信息 DB.DB_LinksToken) DB.DB_Ka {
	var 局_卡详情 DB.DB_Ka
	局_卡详情, err := Ser_Ka.Id取详情(局_在线信息.Uid)
	if err != nil {
		return 局_卡详情
	}
	return 局_卡详情
}
func jS_取软件用户详情(局_在线信息 DB.DB_LinksToken) DB.DB_AppUser {
	var 局_详情 DB.DB_AppUser
	局_详情, ok := Ser_AppUser.Uid取详情(局_在线信息.LoginAppid, 局_在线信息.Uid)
	if ok {
		return 局_详情
	}
	return 局_详情
}
func jS_用户Id增减余额(局_在线信息 DB.DB_LinksToken, 增减值 float64, 原因 string) js对象_通用返回 {
	is增加 := 增减值 >= 0

	新余额, err := Ser_User.Id余额增减(局_在线信息.Uid, Float64取绝对值(增减值), is增加)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}

	go Ser_Log.Log_写余额日志(局_在线信息.User, 局_在线信息.Ip, 原因+"|新余额≈"+Float64到文本(新余额, 2), 增减值)

	return js对象_通用返回{IsOk: true, Err: ""}
}
func jS_用户Id增减积分(局_在线信息 DB.DB_LinksToken, 增减值 float64, 原因 string) js对象_通用返回 {
	is增加 := 增减值 >= 0

	局_AppUserId := Ser_AppUser.User或卡号取Id(局_在线信息.LoginAppid, 局_在线信息.User)
	err := Ser_AppUser.Id积分增减(局_在线信息.LoginAppid, 局_AppUserId, Float64取绝对值(增减值), is增加)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, 局_在线信息.Ip, 原因, 增减值, 局_在线信息.LoginAppid, 1)
	return js对象_通用返回{IsOk: true, Err: ""}
}
func jS_用户Id增减时间点数(AppId int, 局_在线信息 DB.DB_LinksToken, 增减值 int, 原因 string) js对象_通用返回 {
	is增加 := 增减值 >= 0
	//获取增减值的绝对值
	增减值 = S三元(增减值 > 0, 增减值, -增减值)

	局_AppUserId := Ser_AppUser.User或卡号取Id(AppId, 局_在线信息.User)
	if 局_AppUserId == 0 {
		局_AppUserId = Ser_AppUser.Uid取Id(AppId, 局_在线信息.Uid)
	}

	err := Ser_AppUser.Id点数增减(AppId, 局_AppUserId, int64(增减值), is增加)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	if Ser_AppInfo.App是否为计点(AppId) {
		go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, 局_在线信息.Ip, 原因, float64(增减值), AppId, 2)
	} else {
		go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, 局_在线信息.Ip, 原因, float64(增减值), AppId, 3)
	}

	return js对象_通用返回{IsOk: true, Err: ""}
}

type js对象_通用返回 struct {
	IsOk bool        `json:"IsOk"`
	Err  string      `json:"Err"`
	Data interface{} `json:"Data"`
}

func jS_读公共变量(变量名 string) string {

	return publicData.L_publicData.Q取值(&gin.Context{}, 1, 变量名)
}
func jS_置公共变量(变量名, 值 string) bool {
	var err error
	if publicData.L_publicData.Name是否存在(&gin.Context{}, 1, 变量名) {
		err = publicData.L_publicData.Z置值(&gin.Context{}, 1, 变量名, 值)
	} else {
		var 局_新公共变量 = DB.DB_PublicData{AppId: 1, Name: 变量名, Value: 值, Type: 1, IsVip: 0, Time: int(time.Now().Unix()), Note: ""}
		err = publicData.L_publicData.C创建(&gin.Context{}, 局_新公共变量)
	}
	return err == nil
}
func jS_置在线动态标签(局_在线信息 DB.DB_LinksToken, 新动态标签 string) bool {
	return Ser_LinkUser.Set动态标签(局_在线信息.Id, 新动态标签) == nil
}
func jS_网页访问_GET(Url string, 协议头一行一个 string, Cookies string, 超时秒数 int, 代理ip string) js对象_网页响应 {

	if 超时秒数 == 0 {
		超时秒数 = 15
	}

	client := req.C().EnableInsecureSkipVerify().SetTimeout(time.Duration(超时秒数) * time.Second).EnableForceHTTP1()

	if 代理ip != "" {
		client.SetProxyURL(代理ip)
	}
	request := client.R()
	request.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.5735.289 Safari/537.36")
	if Cookies != "" {
		request.SetHeader("Cookie", Cookies)
	}
	var 局_协议头数组 []string
	局_协议头数组 = W文本_分割文本(协议头一行一个, "\r")

	for _, 值 := range 局_协议头数组 {
		if strings.Index(值, ":") != -1 {
			request.SetHeader(W文本_取文本左边(值, ":"), W文本_取文本右边(值, ":"))
		}
	}

	ret, err := request.Get(Url)
	if err != nil {
		return js对象_网页响应{StatusCode: 0, Cookies: "", Headers: "", Body: err.Error()}
	}

	局_响应头信息 := ret.HeaderToString()

	局_临时文本数组 := W文本_分割文本(Cookies, ";") //分割传入的文本
	var 局_临时MAP = make(map[string]string)
	for _, 值 := range 局_临时文本数组 {
		局_临时MAP[W文本_取文本左边(值, "=")] = W文本_取文本右边(值, "=")
	}

	for _, 值 := range ret.Cookies() {
		//如果是重复的 新的会替换掉旧的cookies
		局_临时MAP[值.Name] = 值.Value
	}
	Cookies = ""
	for key, val := range 局_临时MAP {
		if key != "" {
			Cookies += key + "=" + val + ";"
		}
	}
	return js对象_网页响应{StatusCode: ret.StatusCode, Cookies: Cookies, Headers: 局_响应头信息, Body: ret.String(), Base64Body: B编码_BASE64编码(ret.Bytes())}

}
func jS_网页访问_POST(Url, post string, 协议头一行一个 string, Cookies string, 超时秒数 int, 代理ip string) js对象_网页响应 {
	if 超时秒数 == 0 {
		超时秒数 = 15
	}
	client := req.C().EnableInsecureSkipVerify().SetTimeout(time.Duration(超时秒数) * time.Second)

	if 代理ip != "" {
		client.SetProxyURL(代理ip)
	}
	request := client.R()
	request.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.5735.289 Safari/537.36")

	if W文本_是否JSON(post) {
		request.SetHeader("Content-Type", "application/json")
		request.SetHeader("Accept", "application/json, text/plain, */*")
	}

	var 局_协议头数组 []string
	局_协议头数组 = W文本_分割文本(协议头一行一个, "\r")

	for _, 值 := range 局_协议头数组 {
		if strings.Index(值, ":") != -1 {
			request.SetHeader(W文本_取文本左边(值, ":"), W文本_取文本右边(值, ":"))
		}
	}
	if Cookies != "" {
		request.SetHeader("Cookie", Cookies)
	}
	ret, err := request.SetBody(post).Post(Url)
	if err != nil {
		return js对象_网页响应{StatusCode: 0, Cookies: "", Headers: "", Body: err.Error()}
	}

	局_响应头信息 := ret.HeaderToString()

	局_临时文本数组 := W文本_分割文本(Cookies, ";") //分割传入的文本
	var 局_临时MAP = make(map[string]string)
	for _, 值 := range 局_临时文本数组 {
		局_临时MAP[W文本_取文本左边(值, "=")] = W文本_取文本右边(值, "=")
	}

	for _, 值 := range ret.Cookies() {
		//如果是重复的 新的会替换掉旧的cookies
		局_临时MAP[值.Name] = 值.Value
	}
	Cookies = ""
	for key, val := range 局_临时MAP {
		if key != "" {
			Cookies += key + "=" + val + ";"
		}
	}
	return js对象_网页响应{StatusCode: ret.StatusCode, Cookies: Cookies, Headers: 局_响应头信息, Body: ret.String(), Base64Body: B编码_BASE64编码(ret.Bytes())}
}

type js对象_网页响应 struct {
	StatusCode int    `json:"StatusCode"`
	Headers    string `json:"Headers"`
	Cookies    string `json:"Cookies"`
	Body       string `json:"Body"`
	Base64Body string `json:"base64Body"`
}

// 执行sql查询,支持预处理绑定参数
func jS_执行SQL查询(SQL string, data []interface{}) js对象_通用返回 {
	var results []map[string]interface{}

	// 执行 SQL 查询
	if err := global.GVA_DB.Raw(SQL, data...).Scan(&results).Error; err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	// 将查询结果转换为 JSON 格式的字符串
	jsonStr, err := json.Marshal(results)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	fmt.Println(string(jsonStr))
	if results == nil { //防止返回json Null  应该返回空数组
		results = make([]map[string]interface{}, 0)
	}

	return js对象_通用返回{IsOk: true, Err: string(jsonStr), Data: results}

}
func jS_执行SQL功能(SQL string, data []interface{}) js对象_通用返回 {
	局_执行结果 := global.GVA_DB.Exec(SQL, data...)

	if err := 局_执行结果.Error; err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	return js对象_通用返回{IsOk: true, Err: strconv.FormatInt(局_执行结果.RowsAffected, 10)}
}

func jS_任务池_任务创建(局_在线信息 DB.DB_LinksToken, 任务类型ID int, 任务数据 string) js对象_通用返回 {
	//{"Api":"TaskPoolNew","TaskTypeId":1,"Parameter":"{'a':1}","Time":1684752350,"Status":28986}

	局_任务类型, err := Ser_TaskPool.Task类型读取(任务类型ID)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: "任务类型ID不存在"}
	}

	if 局_任务类型.Status != 1 {
		return js对象_通用返回{IsOk: false, Err: "任务类型ID维护中"}
	}

	局_任务数据 := 任务数据 //Parameter
	AppInfo := Ser_AppInfo.App取App详情(局_在线信息.LoginAppid)
	if 局_任务类型.HookSubmitDataStart != "" {

		局_任务数据, _, err = JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 0)
		if err != nil {
			return js对象_通用返回{IsOk: false, Err: err.Error()}
		}
	}

	任务Id, err := Ser_TaskPool.Task数据创建加入队列(局_任务类型.Id, 局_任务数据)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: "Task数据创建加入队列失败"}
	}

	if 局_任务类型.HookSubmitDataEnd != "" {
		局_任务数据, _, err = JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 1)
		if err != nil {
			return js对象_通用返回{IsOk: false, Err: err.Error()}
		}
	}

	//新任务,使用mqtt通知
	if 局_任务类型.MqttTopicMsg != "" {
		局_临时文本 := fmt.Sprintf(`{"taskId":%d,"time":%d}`, 局_任务类型.Id, time.Now().Unix())
		go mqttClient.L_mqttClient.F发送消息(nil, 局_任务类型.MqttTopicMsg, 局_临时文本)
	}

	return js对象_通用返回{IsOk: true, Err: "", Data: gin.H{"TaskUuid": 任务Id}}

}

func jS_任务池_任务查询(任务Uuid string) js对象_通用返回 {

	局_uuid := 任务Uuid
	if len(局_uuid) != 36 { //提前筛选,优化
		return js对象_通用返回{IsOk: false, Err: "任务Uuid错误"}
	}
	局_任务数据, err := Ser_TaskPool.Task数据读取_单条(局_uuid)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: "任务Uuid错误"}
	}

	a := gin.H{"Status": 局_任务数据.Status, "ReturnData": 局_任务数据.ReturnData, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd}

	return js对象_通用返回{IsOk: true, Err: "", Data: a}

}
func jS_置用户云配置(局_在线信息 DB.DB_LinksToken, 配置名称, 配置值 string) js对象_通用返回 {

	if 配置名称 == "" {
		return js对象_通用返回{IsOk: false, Err: "配置名称不能为空"}
	}
	if 局_在线信息.LoginAppid <= 0 {
		return js对象_通用返回{IsOk: false, Err: "登录信息必须大于0"}
	}
	if 局_在线信息.Uid <= 0 {
		return js对象_通用返回{IsOk: false, Err: "Uid必须大于0"}
	}
	if 配置值 == "" { //值为空则删
		global.GVA_DB.Model(DB.DB_UserConfig{}).Delete(DB.DB_UserConfig{
			AppId: 局_在线信息.LoginAppid,
			Uid:   局_在线信息.Uid,
			Name:  配置值,
		})
	} else {
		_ = Ser_UserConfig.Z置值(局_在线信息.LoginAppid, 局_在线信息.Uid, 配置名称, 配置值)
	}

	return js对象_通用返回{IsOk: true, Err: "成功"}
}

func jS_取用户云配置(局_在线信息 DB.DB_LinksToken, 配置名称 string) js对象_通用返回 {

	if 配置名称 == "" {
		return js对象_通用返回{IsOk: false, Err: "配置名称不能为空"}
	}
	if 局_在线信息.LoginAppid <= 0 {
		return js对象_通用返回{IsOk: false, Err: "登录信息必须大于0"}
	}
	if 局_在线信息.Uid <= 0 {
		return js对象_通用返回{IsOk: false, Err: "Uid必须大于0"}
	}
	局_值 := Ser_UserConfig.Q取值(局_在线信息.LoginAppid, 局_在线信息.Uid, 配置名称)

	return js对象_通用返回{IsOk: true, Err: "成功", Data: 局_值}
}

func jS_取缓存(配置名称 string) (ret string) {

	if 配置名称 == "" {
		return
	}
	if 临时数据, ok := global.H缓存.Get("gghsjs_" + 配置名称); ok {
		ret = 临时数据.(string)
	}
	return
}
func jS_置缓存(配置名称, 配置值 string, 有效期 int) bool {
	if 配置名称 == "" {
		return false
	}
	if 配置值 == "" {
		global.H缓存.Delete("gghsjs_" + 配置名称)
	} else {
		global.H缓存.Set("gghsjs_"+配置名称, 配置值, time.Duration(有效期)*time.Second)
	}

	return true
}

func jS_用户名或卡号取uid(应用id int, 用户名或卡号 string) int {

	if Ser_AppInfo.App是否为卡号(应用id) {
		return Ser_Ka.Ka卡号取id(应用id, 用户名或卡号)
	}
	return Ser_User.User用户名取id(用户名或卡号)
}

func jS_置黑名单(AppId int, 黑名单信息, 备注 string) js对象_通用返回 {
	var S = service.S_Blacklist{}
	tx := *global.GVA_DB
	err := S.Create(&tx, db.DB_Blacklist{AppId: AppId, ItemKey: 黑名单信息, Note: 备注})
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	return js对象_通用返回{IsOk: true, Err: "成功"}
}

func jS_mqtt发送消息(主题 string, 消息 string) js对象_通用返回 {

	err := mqttClient.L_mqttClient.F发送消息(nil, 主题, 消息)

	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	return js对象_通用返回{IsOk: true, Err: "成功"}
}

func jS_任务池Uuid添加到队列(uuid string) js对象_通用返回 {

	err := Ser_TaskPool.Uuid_添加到队列(uuid)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	return js对象_通用返回{IsOk: true, Err: "成功"}
}
