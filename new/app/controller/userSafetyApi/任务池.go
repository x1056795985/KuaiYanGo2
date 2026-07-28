package userSafetyApi

import (
	. "EFunc/utils"
	"encoding/json"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Js"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/userSafetyApi/taskPool"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	response2 "server/new/app/models/response"
	"server/new/app/service"
	"strconv"
	"time"
)

func UserApi_任务池_任务创建(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.FailMsg(c, constant.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.FailMsg(c, constant.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.Fail(c, constant.Status_未登录)
		return
	}
	db := *global.GVA_DB
	//{"Api":"TaskPoolNew","TaskTypeId":1,"Parameter":"{'a':1}","Time":1684752350,"Status":28986}
	局_任务类型, err := service.NewTaskPoolType(c, &db).Info(strconv.Itoa(局_ctx.Q请求明文.Get("TaskTypeId").Int()))
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "任务类型Id不存在")
		return
	}
	if 局_任务类型.Status != 1 {
		response.FailMsg(c, constant.Status_操作失败, "维护中")
		return
	}
	局_任务数据 := ""
	if 局_ctx.Q请求明文.Get("Parameter").IsMap() {
		局_任务数据 = 局_ctx.Q请求明文.Get("Parameter").String()
	} else {
		局_任务数据 = 局_ctx.Q请求明文.Get("Parameter").String()
	}
	if 局_任务类型.HookSubmitDataStart != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_任务池Hook处理(c, &局_ctx.AppInfo, &局_ctx.Z在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 0)
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, err.Error())
			return
		}
	}
	任务Id, err := taskPool.L_L_taskPool.Task数据创建加入队列(c, 局_任务类型.Id, 局_任务数据, 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "Task数据创建加入队列失败"+err.Error())
		return
	}
	if 局_任务类型.HookSubmitDataEnd != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_任务池Hook处理(c, &局_ctx.AppInfo, &局_ctx.Z在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 1)
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, err.Error())
			return
		}
	}

	response.OkData(c, gin.H{"TaskUuid": 任务Id})
	return
}

func UserApi_任务池_任务查询(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.Fail(c, constant.Status_未登录)
		return
	}
	db := *global.GVA_DB
	//{"Api":"TaskPoolGetData","TaskUuid":"388f3cb1-ee27-4a5c-979d-a17cf3107dcd","Time":1684761030,"Status":12622}

	局_uuid := 局_ctx.Q请求明文.Get("TaskUuid").String()
	if len(局_uuid) != 36 { //提前筛选,优化
		response.FailMsg(c, constant.Status_操作失败, "任务Uuid错误")
		return
	}
	局_任务数据, err := service.NewTaskPoolData(c, &db).Info(局_uuid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "任务Uuid错误")
		return
	}
	var mapkv map[string]interface{}

	//局_任务数据.ReturnData 判断字符串是否为json格式如果是json则解析
	if json.Unmarshal([]byte(局_任务数据.ReturnData), &mapkv) == nil {
		response.OkData(c, gin.H{"Status": 局_任务数据.Status, "ReturnData": mapkv, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd})
	} else {
		response.OkData(c, gin.H{"Status": 局_任务数据.Status, "ReturnData": 局_任务数据.ReturnData, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd})
	}
	return
}

// 1.0.326+版本添加可用
func UserApi_任务池_取任务列表(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.Fail(c, constant.Status_未登录)
		return
	}
	db := *global.GVA_DB
	//{"Api":"TaskPoolGetDataList","Page":1,"Order":1,"Size":30,"Tid":1,"isSimple":1,"Time":1684761030,"Status":12622}
	var 请求 = request.List{
		Page:     局_ctx.Q请求明文.Get("Page").Int(),
		Size:     局_ctx.Q请求明文.Get("Size").Int(),
		Type:     0,
		Keywords: "",
		Order:    局_ctx.Q请求明文.Get("Order").Int(), // 0 倒序 1 正序
	}
	i, list, err := service.NewTaskPoolData(c, &db).GetList(请求, 局_ctx.Q请求明文.Get("Tid").Int(), 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	var list2 = make([]struct {
		dbm.DB_TaskPoolData
		Msg string `json:"Msg"`
	}, len(list))

	for a, _ := range list {
		list2[a].DB_TaskPoolData = list[a]
		// 取出 json msg 部分信息当做失败提示  ,"msg":"主页_角色详细信息卡片失败",  取文本中间吧不解析json了,提高响应速度
		list2[a].Msg = W文本_取出中间文本(list[a].ReturnData, `"msg"`, `",`)

		if 局_ctx.Q请求明文.Get("isSimple").Int() == 1 { //简略信息,节省网络通讯时间
			list[a].SubmitData = ""
			list[a].ReturnData = ""
		}
	}
	response.OkData(c, response2.GetList{List: list, Count: i})
	return
}
func UserApi_任务池_任务处理获取(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.Fail(c, constant.Status_未登录)
		return
	}

	// {"Api":"TaskPoolGetTask","GetTaskNumber":5,"GetTaskTypeId":[1],"Time":1684764215,"Status":18042}
	局_最大数量 := 局_ctx.Q请求明文.Get("GetTaskNumber").Int()

	var 局_可获取任务类型ID = make([]int, 局_ctx.Q请求明文.Len("GetTaskTypeId"))
	for v := range 局_ctx.Q请求明文.Len("GetTaskTypeId") {
		局_可获取任务类型ID[v] = 局_ctx.Q请求明文.Get("GetTaskTypeId." + D到文本(v)).Int()
	}
	局_任务UUID := taskPool.L_L_taskPool.Task队列弹出任务(局_可获取任务类型ID, 局_最大数量, 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.Uid)
	var 局_已获取任务数据 []dbm.TaskPool_数据_精简
	if len(局_任务UUID) > 0 {
		局_已获取任务数据 = taskPool.L_L_taskPool.Task数据读取_数组(局_任务UUID)
	} else {
		局_已获取任务数据 = []dbm.TaskPool_数据_精简{}
	}

	response.OkData(c, 局_已获取任务数据)
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
				response.FailMsg(c, constant.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.FailMsg(c, constant.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.Fail(c, constant.Status_未登录)
		return
	}
	db := *global.GVA_DB
	//{"Api":"TaskPoolSetTask","TaskUuid":"f2e87ec0-4e0a-404d-a374-124d553a5a35","TaskStatus":40160,"TaskReturnData":"BB6CB5C68DF4652941CAF652A366F2D8","Time":1684769068}

	局_uuid := 局_ctx.Q请求明文.Get("TaskUuid").String()
	if len(局_uuid) != 36 { //提前筛选,优化
		response.FailMsg(c, constant.Status_操作失败, "任务Uuid错误")
		return
	}

	局_Tid := taskPool.L_L_taskPool.Task数据读取Tid(&db, 局_uuid)

	局_任务类型, err := service.NewTaskPoolType(c, &db).Info(strconv.Itoa(局_Tid))
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "该UUID的任务类型Id不存在")
		return
	}
	局_任务数据 := 局_ctx.Q请求明文.Get("TaskReturnData").String()
	局_任务状态 := 局_ctx.Q请求明文.Get("TaskStatus").Int()
	if 局_任务类型.HookReturnDataStart != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_任务池Hook处理(c, &局_ctx.AppInfo, &局_ctx.Z在线信息, 局_任务类型.HookReturnDataStart, 局_任务数据, 局_任务状态)
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, err.Error())
			return
		}
	}

	// 数据修改 Status=0 或ReturnData="" 不修改
	局_UpData := make(map[string]interface{}, 3)
	局_UpData["TimeEnd"] = time.Now().Unix()
	if 局_任务状态 != 0 {
		局_UpData["Status"] = 局_任务状态
	}
	if 局_任务数据 != "" {
		局_UpData["ReturnData"] = 局_任务数据
	}
	_, err = service.NewTaskPoolData(c, &db).Update(局_uuid, 局_UpData)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "任务数据写入数据库失败")
		return
	}

	if 局_任务类型.HookReturnDataEnd != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_任务池Hook处理(c, &局_ctx.AppInfo, &局_ctx.Z在线信息, 局_任务类型.HookReturnDataEnd, 局_任务数据, 局_任务状态)
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, err.Error())
			return
		}
	}

	response.Ok(c)
	return
}
func UserApi_任务池_取类型状态(c *gin.Context) {
	/*	var AppInfo dbm.DB_AppInfo
		var 局_在线信息 dbm.DB_LinksToken
		局_ctx := 取上下文(c)
		if !检测用户登录在线正常(&局_ctx.Z在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
			response.Fail(c, response.Status_未登录)
			return
		}*/

	//{"Api":"TaskPoolGetTypeStatus","Time":1684769068}
	var DB_TaskPool_类型 []dbm.TaskPool_类型
	_ = global.GVA_DB.Model(dbm.TaskPool_类型{}).Select("Id,Status").Find(&DB_TaskPool_类型).Error
	var 局_map = make(map[string]int, len(DB_TaskPool_类型))
	for _, v := range DB_TaskPool_类型 {
		局_map["id"+strconv.Itoa(v.Id)] = v.Status
	}
	response.OkData(c, 局_map)
	return
}
