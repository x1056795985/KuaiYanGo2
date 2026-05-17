package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

// LogLogin 登录日志
type LogLogin struct {
	Common.Common
}

func NewLogLoginController() *LogLogin {
	return &LogLogin{}
}

// Info 获取详情
func (C *LogLogin) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogLogin{}
	tx := *global.GVA_DB
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

type 请求_LogLoginGetList struct {
	request.List
	Appid int `json:"appid"` // 登录类型筛选
}

// GetList 登录日志列表
func (C *LogLogin) GetList(c *gin.Context) {
	var 请求 请求_LogLoginGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogLogin{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetList(&tx, 请求.List, 请求.Appid)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	// 附加应用名称
	AppNameMap := S.GetAppNameMap(dataList)
	response.OkWithDetailed(struct {
		List    interface{}       `json:"list"`
		Count   int64             `json:"count"`
		AppName map[string]string `json:"appName"`
	}{dataList, 总数, AppNameMap}, "获取成功", c)
}

type 请求_LogLoginBatchDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`     // 1删除ID数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前 7关键字
	Keywords string `json:"keywords"`
}

// Delete 批量删除
func (C *LogLogin) Delete(c *gin.Context) {
	var 请求 请求_LogLoginBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogLogin{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDelete(&tx, 请求.Id, 请求.Type, 请求.Keywords)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}
