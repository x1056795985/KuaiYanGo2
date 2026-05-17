package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

// LogMoney 余额日志
type LogMoney struct {
	Common.Common
}

func NewLogMoneyController() *LogMoney {
	return &LogMoney{}
}

// Info 获取详情
func (C *LogMoney) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogMoney{}
	tx := *global.GVA_DB
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

// GetList 余额日志列表
func (C *LogMoney) GetList(c *gin.Context) {
	var 请求 request.List
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogMoney{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetList(&tx, 请求)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "获取成功", c)
}

type 请求_LogMoneyBatchDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

// Delete 批量删除
func (C *LogMoney) Delete(c *gin.Context) {
	var 请求 请求_LogMoneyBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogMoney{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDelete(&tx, 请求.Id, 请求.Type, 请求.Keywords)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}
