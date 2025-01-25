package UserApi

import (
	. "EFunc/utils"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Ser_Js"
	"server/Service/Ser_TaskPool"
	"server/api/UserApi/response"
	"server/global"
	"server/new/app/logic/common/cloudStorage"
	"server/new/app/logic/common/mqttClient"
	"server/new/app/models/request"
	response2 "server/new/app/models/response"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

func UserApi_任务池_任务创建(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolNew","TaskTypeId":1,"Parameter":"{'a':1}","Time":1684752350,"Status":28986}
	局_任务类型, err := Ser_TaskPool.Task类型读取(请求json.GetInt("TaskTypeId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务类型Id不存在")
		return
	}
	if 局_任务类型.Status != 1 {
		response.X响应状态消息(c, response.Status_操作失败, "维护中")
		return
	}
	局_任务数据 := ""
	if 请求json.Get("Parameter").Type().String() == "object" {
		局_任务数据 = 请求json.Get("Parameter").String()
	} else {
		局_任务数据 = string(请求json.GetStringBytes("Parameter"))
	}
	if 局_任务类型.HookSubmitDataStart != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 0)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}
	任务Id, err := Ser_TaskPool.Task数据创建加入队列(局_任务类型.Id, 局_任务数据, 局_在线信息.LoginAppid, 局_在线信息.Uid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "Task数据创建加入队列失败")
		return
	}
	if 局_任务类型.HookSubmitDataEnd != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 1)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}
	//新任务,使用mqtt通知
	if 局_任务类型.MqttTopicMsg != "" {
		局_临时文本 := fmt.Sprintf(`{"taskId":%d,"time":%d}`, 局_任务类型.Id, time.Now().Unix())
		//因为有网络通讯单开协程处理,不能卡请求耗时
		go mqttClient.L_mqttClient.F发送消息(nil, 局_任务类型.MqttTopicMsg, 局_临时文本)
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"TaskUuid": 任务Id})
	return
}

func UserApi_任务池_任务查询(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolGetData","TaskUuid":"388f3cb1-ee27-4a5c-979d-a17cf3107dcd","Time":1684761030,"Status":12622}

	局_uuid := string(请求json.GetStringBytes("TaskUuid"))
	if len(局_uuid) != 36 { //提前筛选,优化
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}
	局_任务数据, err := Ser_TaskPool.Task数据读取_单条(局_uuid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}
	var mapkv map[string]interface{}

	//局_任务数据.ReturnData 判断字符串是否为json格式如果是json则解析
	if json.Unmarshal([]byte(局_任务数据.ReturnData), &mapkv) == nil {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 局_任务数据.Status, "ReturnData": mapkv, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd})
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 局_任务数据.Status, "ReturnData": 局_任务数据.ReturnData, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd})
	}
	return
}

// 1.0.326+版本添加可用
func UserApi_任务池_取任务列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolGetDataList","Page":1,"Order":1,"Size":30,"Tid":1,"Time":1684761030,"Status":12622}
	db := *global.GVA_DB
	var 请求 = request.List{
		Page:     请求json.GetInt("Page"),
		Size:     请求json.GetInt("Size"),
		Type:     0,
		Keywords: "",
		Order:    请求json.GetInt("Order"), // 0 倒序 1 正序
	}
	i, list, err := service.NewTaskPoolData(c, &db).GetList(请求, 请求json.GetInt("Tid"), 局_在线信息.LoginAppid, 局_在线信息.Uid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), response2.GetList{List: list, Count: i})
	return
}
func UserApi_任务池_任务处理获取(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"TaskPoolGetTask","GetTaskNumber":5,"GetTaskTypeId":[1],"Time":1684764215,"Status":18042}
	局_最大数量 := 请求json.GetInt("GetTaskNumber")
	局_临时 := 请求json.GetArray("GetTaskTypeId")
	var 局_可获取任务类型ID = make([]int, len(局_临时))
	for 索引, _ := range 局_临时 {
		局_可获取任务类型ID[索引], _ = 局_临时[索引].Int()
	}
	局_任务UUID := Ser_TaskPool.Task队列弹出任务(局_可获取任务类型ID, 局_最大数量, 局_在线信息.LoginAppid, 局_在线信息.Uid)
	var 局_已获取任务数据 []DB.TaskPool_数据_精简
	if len(局_任务UUID) > 0 {
		局_已获取任务数据 = Ser_TaskPool.Task数据读取_数组(局_任务UUID)
	} else {
		局_已获取任务数据 = []DB.TaskPool_数据_精简{}
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_已获取任务数据)
	return
}

type TaskPool_数据_精简 struct {
	Uuid string `json:"uuid" gorm:"column:uuid;size:36;primarykey;"`
	//LId        int    `json:"LId" gorm:"column:LId;comment:在线id,只允许相同的查询任务"` 直接用UUid,不可能重复的除了获取者别人也猜不到ID
	Tid        int    `json:"Tid" gorm:"column:Tid;comment:对应的任务类型Id"`
	TimeStart  int    `json:"TimeStart" gorm:"column:TimeStart;comment:任务创建时间戳"`
	TimeEnd    int    `json:"TimeEnd" gorm:"column:TimeEnd;comment:任务结束时间戳"`
	SubmitData string `json:"SubmitData" gorm:"column:SubmitData;comment:生产提交数据"`
	ReturnData string `json:"ReturnData" gorm:"column:ReturnData;comment:消费返回数据"`
	Status     int    `json:"Status" gorm:"column:Status;comment:任务状态,"` //1 已创建,2任务处理中,3成功,4任务失败
}

func UserApi_任务池_任务处理返回(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolSetTask","TaskUuid":"f2e87ec0-4e0a-404d-a374-124d553a5a35","TaskStatus":40160,"TaskReturnData":"BB6CB5C68DF4652941CAF652A366F2D8","Time":1684769068}

	局_uuid := string(请求json.GetStringBytes("TaskUuid"))
	if len(局_uuid) != 36 { //提前筛选,优化
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}

	局_Tid := Ser_TaskPool.Task数据读取Tid(局_uuid)

	局_任务类型, err := Ser_TaskPool.Task类型读取(局_Tid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "该UUID的任务类型Id不存在")
		return
	}
	局_任务数据 := string(请求json.GetStringBytes("TaskReturnData"))
	局_任务状态 := 请求json.GetInt("TaskStatus")
	if 局_任务类型.HookReturnDataStart != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookReturnDataStart, 局_任务数据, 局_任务状态)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	err = Ser_TaskPool.Task数据修改(局_uuid, 局_任务状态, 局_任务数据)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务数据写入数据库失败")
		return
	}

	if 局_任务类型.HookReturnDataEnd != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_任务池Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookReturnDataEnd, 局_任务数据, 局_任务状态)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}
func UserApi_任务池_取类型状态(c *gin.Context) {
	/*	var AppInfo DB.DB_AppInfo
		var 局_在线信息 DB.DB_LinksToken
		Y用户数据信息还原(c, &AppInfo, &局_在线信息)
		if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
			response.X响应状态(c, response.Status_未登录)
			return
		}*/

	//{"Api":"TaskPoolGetTypeStatus","Time":1684769068}
	var DB_TaskPool_类型 []DB.TaskPool_类型
	_ = global.GVA_DB.Model(DB.TaskPool_类型{}).Select("Id,Status").Find(&DB_TaskPool_类型).Error
	var 局_map = make(map[string]int, len(DB_TaskPool_类型))
	for _, v := range DB_TaskPool_类型 {
		局_map["id"+strconv.Itoa(v.Id)] = v.Status
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_map)
	return
}

// 1.0.325+版本添加可用
func UserApi_云存储_取文件上传授权(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetUploadToken","Path":"8987657"}
	path := strings.TrimSpace(string(请求json.GetStringBytes("Path")))

	if path == "" || strings.Index(path, ".") == -1 || W文本_取右边(path, 1) == "/" {
		response.X响应状态消息(c, response.Status_操作失败, "暂不支持该文件类型")
		return
	}
	取文件上传授权, err := cloudStorage.L_云存储.Q取文件上传授权(c, path)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 取文件上传授权)
	return
}
