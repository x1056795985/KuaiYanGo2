package Ser_Js

// https://blog.csdn.net/wyongqing/article/details/124704136   参考地址
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	. "github.com/duolabmeng6/efun/efun"
	E "github.com/duolabmeng6/goefun/eTool"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicData"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	"server/utils"
	"strconv"
	"time"
)

func JS引擎初始化_用户(AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken) *goja.Runtime {
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

	return vm

}
func JS引擎初始化_Hook处理(AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken, Hook函数, 任务数据 string, 局_任务状态 int) (string, int, error) {
	局_PublicJs, err := Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_Hook函数, Hook函数)
	if err != nil {
		return "", 局_任务状态, err
	}

	vm := JS引擎初始化_用户(AppInfo, 在线信息)
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

	新余额, err := Ser_User.Id余额增减(局_在线信息.Uid, utils.Float64取绝对值(增减值), is增加)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}

	go Ser_Log.Log_写余额日志(局_在线信息.User, 局_在线信息.Ip, 原因+"|新余额≈"+utils.Float64到文本(新余额, 2), 增减值)

	return js对象_通用返回{IsOk: true, Err: ""}
}
func jS_用户Id增减积分(局_在线信息 DB.DB_LinksToken, 增减值 float64, 原因 string) js对象_通用返回 {
	is增加 := 增减值 >= 0

	局_AppUserId := Ser_AppUser.User或卡号取Id(局_在线信息.LoginAppid, 局_在线信息.User)
	err := Ser_AppUser.Id积分增减(局_在线信息.LoginAppid, 局_AppUserId, utils.Float64取绝对值(增减值), is增加)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, 局_在线信息.Ip, 原因, 增减值, 局_在线信息.LoginAppid, 1)
	return js对象_通用返回{IsOk: true, Err: ""}
}
func jS_用户Id增减时间点数(AppId int, 局_在线信息 DB.DB_LinksToken, 增减值 int, 原因 string) js对象_通用返回 {
	is增加 := 增减值 >= 0
	局_AppUserId := Ser_AppUser.User或卡号取Id(局_在线信息.LoginAppid, 局_在线信息.User)
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
	IsOk bool                   `json:"IsOk"`
	Err  string                 `json:"Err"`
	Data map[string]interface{} `json:"Data"`
}

func jS_读公共变量(变量名 string) string {
	return Ser_PublicData.P取值(1, 变量名)
}
func jS_置公共变量(变量名, 值 string) bool {
	var err error
	if Ser_PublicData.Name是否存在(1, 变量名) {
		err = Ser_PublicData.P置值(1, 变量名, 值)
	} else {

		var 局_新公共变量 = DB.DB_PublicData{AppId: 1, Name: 变量名, Value: 值, Type: 1, IsVip: 0, Time: int(time.Now().Unix()), Note: ""}
		err = global.GVA_DB.Model(DB.DB_PublicData{}).Create(&局_新公共变量).Error
	}
	return err == nil
}
func jS_置在线动态标签(局_在线信息 DB.DB_LinksToken, 新动态标签 string) bool {
	return Ser_LinkUser.Set动态标签(局_在线信息.Id, 新动态标签) == nil
}
func jS_网页访问_GET(Url, 协议头一行一个, Cookies string, 超时秒数 int, 代理ip string) js对象_网页响应 {

	ehttp := NewHttp()
	ehttp.E设置自动管理cookie(utils.W网页_取域名(Url))
	ehttp.E设置超时时间(超时秒数)
	if 代理ip != "" {
		ehttp.E设置全局HTTP代理(代理ip)
	}
	ehttp.E设置全局头信息(协议头一行一个)

	ret, _ := ehttp.Get(Url)
	局_响应头信息 := ehttp.E取所有头信息()

	局_临时文本数组 := E分割文本(Cookies, ";")
	var 局_临时MAP = make(map[string]string)

	for _, 值 := range 局_临时文本数组 {
		局_临时MAP[E.E文本_取左边(值, "=")] = E.E文本_取右边(值, "=")
	}

	for _, 值 := range ehttp.Cookies.Entries() {
		for _, 值2 := range 值 {
			//如果是重复的 新的会替换掉旧的cookies
			局_临时MAP[值2.Name] = 值2.Value
		}
	}
	Cookies = ""
	for key, val := range 局_临时MAP {
		if key != "" {
			Cookies += key + "=" + val + ";"
		}
	}

	return js对象_网页响应{StatusCode: ehttp.E取状态码(), Cookies: Cookies, Headers: 局_响应头信息, Body: ret}

}
func jS_网页访问_POST(Url, post, 协议头一行一个, Cookies string, 超时秒数 int, 代理ip string) js对象_网页响应 {

	ehttp := NewHttp()
	ehttp.E设置自动管理cookie(utils.W网页_取域名(Url))
	ehttp.E设置超时时间(超时秒数)
	if 代理ip != "" {
		ehttp.E设置全局HTTP代理(代理ip)
	}
	ehttp.E设置全局头信息(协议头一行一个)

	ret, _ := ehttp.Post(Url, post)
	局_响应头信息 := ehttp.E取所有头信息()

	局_临时文本数组 := E分割文本(Cookies, ";")
	var 局_临时MAP = make(map[string]string)

	for _, 值 := range 局_临时文本数组 {
		局_临时MAP[E.E文本_取左边(值, "=")] = E.E文本_取右边(值, "=")
	}

	for _, 值 := range ehttp.Cookies.Entries() {
		for _, 值2 := range 值 {
			//如果是重复的 新的会替换掉旧的cookies
			局_临时MAP[值2.Name] = 值2.Value
		}
	}
	Cookies = ""
	for key, val := range 局_临时MAP {
		if key != "" {
			Cookies += key + "=" + val + ";"
		}

	}

	return js对象_网页响应{StatusCode: ehttp.E取状态码(), Cookies: Cookies, Headers: 局_响应头信息, Body: ret}

}

type js对象_网页响应 struct {
	StatusCode int    `json:"StatusCode"`
	Headers    string `json:"Headers"`
	Cookies    string `json:"Cookies"`
	Body       string `json:"Body"`
}

func jS_执行SQL查询(SQL string) js对象_通用返回 {
	var results []map[string]interface{}
	// 执行 SQL 查询
	if err := global.GVA_DB.Raw(SQL).Scan(&results).Error; err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	// 将查询结果转换为 JSON 格式的字符串
	jsonStr, err := json.Marshal(results)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: err.Error()}
	}
	fmt.Println(string(jsonStr))
	return js对象_通用返回{IsOk: true, Err: string(jsonStr)}
}
func jS_执行SQL功能(SQL string) js对象_通用返回 {
	局_执行结果 := global.GVA_DB.Exec(SQL)

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

		局_任务数据, _, err = JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 0)
		if err != nil {
			return js对象_通用返回{IsOk: false, Err: err.Error()}
		}
	}

	任务Id, err := Ser_TaskPool.Task数据创建加入队列(局_任务类型.Id, 局_任务数据)
	if err != nil {
		return js对象_通用返回{IsOk: false, Err: "Task数据创建加入队列失败"}
	}

	if 局_任务类型.HookSubmitDataEnd != "" {
		局_任务数据, _, err = JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 1)
		if err != nil {
			return js对象_通用返回{IsOk: false, Err: err.Error()}
		}
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
