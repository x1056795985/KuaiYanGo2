package WebApi

import (
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Ser_Js"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_TaskPool"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

func R任务池_任务处理获取(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"GetTaskNumber":5,"GetTaskTypeId":[1]}
	局_最大数量 := 请求json.GetInt("GetTaskNumber")
	局_临时 := 请求json.GetArray("GetTaskTypeId")
	var 局_可获取任务类型ID = make([]int, len(局_临时))
	for 索引, _ := range 局_临时 {
		局_可获取任务类型ID[索引], _ = 局_临时[索引].Int()
	}
	局_任务UUID := Ser_TaskPool.Task队列弹出任务(局_可获取任务类型ID, 局_最大数量)
	var 局_已获取任务数据 []DB.TaskPool_数据_精简
	if len(局_任务UUID) > 0 {
		局_已获取任务数据 = Ser_TaskPool.Task数据读取_数组(局_任务UUID)
	} else {
		response.OkWithDetailed([]gin.H{}, "获取成功", c)
		return
	}
	response.OkWithDetailed(局_已获取任务数据, "获取成功", c)
	return
}
func R任务池_任务处理返回(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"TaskUuid":"f2e87ec0-4e0a-404d-a374-124d553a5a35","TaskStatus":3,"TaskReturnData":"BB6CB5C68DF4652941CAF652A366F2D8","Time":1684769068}

	局_uuid := string(请求json.GetStringBytes("TaskUuid"))
	if len(局_uuid) != 36 { //提前筛选,优化
		response.FailWithMessage("UUid错误", c)
		return
	}
	局_Tid := Ser_TaskPool.Task数据读取Tid(局_uuid)

	局_任务类型, err := Ser_TaskPool.Task类型读取(局_Tid)
	if err != nil {
		response.FailWithMessage("该UUID的任务类型Id不存在", c)
		return
	}
	局_任务数据 := string(请求json.GetStringBytes("TaskReturnData"))
	局_任务状态 := 请求json.GetInt("TaskStatus")
	if 局_任务类型.HookReturnDataStart != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 局_任务状态)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}

	err = Ser_TaskPool.Task数据修改(局_uuid, 局_任务状态, 局_任务数据)
	if err != nil {
		response.FailWithMessage("任务数据写入数据库失败", c)
		return
	}

	if 局_任务类型.HookReturnDataEnd != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 局_任务状态)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}

	response.Ok(c)
	return
}
func RunJs(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.FailWithMessage("异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String(), c)
			} else {
				response.FailWithMessage("异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息", c)
			}
			return
		}
	}()
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Parameter":"{'a':1}","JsName":"获取用户相关信息"}
	局_耗时 := time.Now().UnixMilli()
	vm := Ser_Js.JS引擎初始化_用户(&AppInfo, &局_在线信息)
	var 局_PublicJs DB.DB_PublicJs
	var err error
	局_PublicJs, err = Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_公共函数, string(请求json.GetStringBytes("JsName")))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	局_云函数型参数 := string(请求json.GetStringBytes("Parameter"))
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.FailWithMessage("JS代码运行失败:"+局_详细错误.String(), c)
		return
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.FailWithMessage("Js中没有["+局_PublicJs.Name+"()]函数", c)
		return
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.FailWithMessage("Js绑定函数到变量失败", c)
		return
	}
	局_return := 局_待执行js函数名(局_云函数型参数)
	response.OkWithDetailed(局_return, "运行成功,耗时:"+strconv.Itoa(int(time.Now().UnixMilli()-局_耗时)), c)
	return
}
