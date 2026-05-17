package controller

import (
	. "EFunc/utils"
	json2 "encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	Db服务 "server/Service/Ser_AppInfo"
	"server/Service/Ser_Js"
	"server/Service/Ser_PublicJs"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

type PublicJsCtrl struct {
	Common.Common
}

func NewPublicJsController() *PublicJsCtrl {
	return &PublicJsCtrl{}
}

type 请求_PublicJsGetInfo struct {
	Name string `json:"name"`
}

type 请求_PublicJsGetList struct {
	AppId    int    `json:"appId"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
}

type 请求_PublicJsDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

type 响应_PublicJsGetList struct {
	List  []响应_PublicJs扩展 `json:"list"`
	Count int64              `json:"count"`
}

type 响应_PublicJs扩展 struct {
	DB.DB_PublicJs
	AppName string `json:"appName"`
}

type 请求_PublicJsSetVip struct {
	Id    []int `json:"id"`
	IsVip int   `json:"isVip"`
}

// Info 获取公共函数详情
func (C *PublicJsCtrl) Info(c *gin.Context) {
	var 请求 请求_PublicJsGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_PublicJs DB.DB_PublicJs
	err := global.GVA_DB.Model(DB.DB_PublicJs{}).Where("Name= ?", 请求.Name).First(&DB_PublicJs).Error
	if err != nil {
		response.FailWithMessage("获取公共变量失败,可能联合主键不存在", c)
		return
	}

	if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value) {
		DB_PublicJs.Value = string(W文件_读入文件(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value))
	} else {
		DB_PublicJs.Value = DB_PublicJs.Value + "[js文件读取失败可能被删除]"
	}

	response.OkWithDetailed(DB_PublicJs, "获取成功", c)
}

// GetPublicAppList 获取公共函数应用列表
func (C *PublicJsCtrl) GetPublicAppList(c *gin.Context) {
	var 局_appid []int
	_ = global.GVA_DB.Model(DB.DB_PublicJs{}).Select("AppId").Group("AppId").Find(&局_appid).Error

	var AppName = Db服务.AppInfo取map列表Int(false)

	type name struct {
		AppId   int    `json:"appId"`
		AppName string `json:"appName"`
	}
	var 局_arr = make([]name, 0, len(局_appid)+4)
	for 索引 := range 局_appid {
		if 局_appid[索引] < 10000 {
			continue
		}
		局_arr = append(局_arr, name{AppId: 局_appid[索引], AppName: AppName[局_appid[索引]]})
	}
	局_arr = append(局_arr, name{1, "全局"})
	局_arr = append(局_arr, name{2, "任务池Hook"})
	局_arr = append(局_arr, name{3, "ApiHook"})
	局_arr = append(局_arr, name{11, "webSocket"})

	response.OkWithDetailed(局_arr, "ok", c)
}

// GetList 获取公共函数列表
func (C *PublicJsCtrl) GetList(c *gin.Context) {
	var 请求 请求_PublicJsGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_PublicJs{})
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else if 请求.Order == 2 {
		局_DB.Order("Id DESC")
	}
	if 请求.AppId > 0 {
		局_DB.Where("AppId = ?", 请求.AppId)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Name LIKE ?", "%"+请求.Keywords+"%")
		}
	}

	var DB_PublicJs []响应_PublicJs扩展
	var 总数 int64
	err := 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Omit("AppName").Find(&DB_PublicJs).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("PublicJsGetList:" + err.Error())
		return
	}

	var AppName = Db服务.App取map列表String(false)
	AppName["1"] = "全局"
	AppName["2"] = "任务池Hook"
	AppName["3"] = "ApiHook"
	AppName["11"] = "webSocket"

	for 索引 := range DB_PublicJs {
		DB_PublicJs[索引].AppName = AppName[strconv.Itoa(DB_PublicJs[索引].AppId)]
	}

	response.OkWithDetailed(响应_PublicJsGetList{DB_PublicJs, 总数}, "获取成功", c)
}

// Delete 批量删除公共函数
func (C *PublicJsCtrl) Delete(c *gin.Context) {
	var 请求 请求_PublicJsDelete
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 && 请求.Type == 1 {
		response.FailWithMessage("数组为空", c)
		return
	}

	for _, 值 := range 请求.Id {
		var DB_PublicJs DB.DB_PublicJs
		err := global.GVA_DB.Model(DB.DB_PublicJs{}).Where("Id = ? ", 值).First(&DB_PublicJs).Error
		if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value) {
			_ = W文件_删除(global.GVA_CONFIG.Q取运行目录 + DB_PublicJs.Value)
			if err != nil {
				fmt.Printf("E删除文件失败%v", err.Error())
			}
		}
	}

	var 影响行数 int64
	var db = global.GVA_DB
	switch 请求.Type {
	case 1:
		if 请求.Type == 1 && len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_PublicJs{}).RowsAffected
	}
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// SaveInfo 保存公共函数信息
func (C *PublicJsCtrl) SaveInfo(c *gin.Context) {
	var 请求 DB.DB_PublicJs
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("变量名不能为空", c)
		return
	}

	var 局_临时Id = Ser_PublicJs.Name取Id([]int{Ser_PublicJs.Js类型_公共函数, Ser_PublicJs.Js类型_任务池Hook函数}, 请求.Name)
	if 局_临时Id != 0 && 局_临时Id != 请求.Id {
		response.FailWithMessage("变量名已存在", c)
		return
	}
	if !Ser_PublicJs.Id是否存在(请求.Id) {
		response.FailWithMessage("变量不存在", c)
		return
	}

	err := Ser_PublicJs.Z置值2(请求)
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// New 新建公共函数
func (C *PublicJsCtrl) New(c *gin.Context) {
	var 请求 DB.DB_PublicJs
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("公共函数名不能为空", c)
		return
	}
	if W文本_是否为数字(请求.Name) {
		response.FailWithMessage("公共函数名不能为纯数字", c)
		return
	}
	if strings.Index(请求.Name, "$api_") != -1 {
		response.FailWithMessage("公共函数名不能包含$api_", c)
		return
	}
	if 请求.Type < 1 {
		response.FailWithMessage("公共函数类型错误", c)
		return
	}

	var 局_临时Id = Ser_PublicJs.Name取Id([]int{Ser_PublicJs.Js类型_公共函数, Ser_PublicJs.Js类型_任务池Hook函数, Ser_PublicJs.Js类型_ApiHook函数}, 请求.Name)
	if 局_临时Id != 0 && 局_临时Id != 请求.Id {
		response.FailWithMessage("公共函数名已存在", c)
		return
	}
	局_禁止字符串 := []string{"\\", "|", ":", "\"", "<", ">", "@", "&", "^", "%", "$", "#", "!", "`", "~", " "}
	W文本_是否存在_任意(请求.Name, 局_禁止字符串)
	if W文本_是否包含关键字(请求.Name, "/") || W文本_是否包含关键字(请求.Name, ".") {
		response.FailWithMessage("函数名不能包含"+strings.Join(局_禁止字符串, ",")+"符号", c)
		return
	}

	err := Ser_PublicJs.C创建(请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// SetVipLimit 批量修改vip限制
func (C *PublicJsCtrl) SetVipLimit(c *gin.Context) {
	var 请求 请求_PublicJsSetVip
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.IsVip < 0 {
		response.FailWithMessage("IsVip值错误", c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("公共变量数组为空", c)
		return
	}

	err := Ser_PublicJs.P批量修改IsVip(请求.Id, 请求.IsVip)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}

// TestExec 测试执行公共函数
func (C *PublicJsCtrl) TestExec(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.FailWithMessage("异常:可能JS函数传参或返回值类型错误,具体:"+局_GoJa错误.String(), c)
			} else {
				response.FailWithMessage("异常:可能JS函数传参或返回值类型错误,具体:js引擎未返回报错信息", c)
			}
			return
		}
	}()

	if !Ser_PublicJs.Id是否存在(请求.Id) {
		response.FailWithMessage("JS公共函数不存在", c)
		return
	}
	局_耗时 := time.Now().UnixMilli()

	var 局_PublicJs DB.DB_PublicJs
	var err error
	局_PublicJs, err = Ser_PublicJs.Q取值2(请求.Id)
	if err != nil {
		response.FailWithMessage("JS公共函数不存在", c)
		return
	}

	if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value) {
		局_PublicJs.Value = string(W文件_读入文件(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value))
	} else {
		response.FailWithMessage("js文件读取失败可能被删除", c)
		return
	}

	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken

	vm := Ser_Js.JS引擎初始化_用户(c, &AppInfo, &局_在线信息, &局_PublicJs)

	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.FailWithMessage("JS代码运行失败:"+局_详细错误.String(), c)
		return
	}
	var 局_待执行js函数名 func(interface{}) string
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

	var 局_返回 string
	if 局_PublicJs.AppId == constant.APPID_WebSocket {
		局_返回 = 局_待执行js函数名(map[string]interface{}{"aaa": 11, "bbb": "22", "cc": []string{"111", "222"}})
	} else {
		局_返回 = 局_待执行js函数名("{}")
	}
	局_耗时 = time.Now().UnixMilli() - 局_耗时
	var mapkv map[string]interface{}

	if W文本_可能为json(局_返回) && 局_PublicJs.AppId == constant.APPID_WebSocket && json2.Unmarshal([]byte(局_返回), &mapkv) == nil {
		response.OkWithDetailed(gin.H{"Return": mapkv, "Time": 局_耗时}, "执行成功,耗时:"+strconv.FormatInt(局_耗时, 10), c)
	} else {
		response.OkWithDetailed(gin.H{"Return": 局_返回, "Time": 局_耗时}, "执行成功,耗时:"+strconv.FormatInt(局_耗时, 10), c)
	}
}
