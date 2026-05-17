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

// AgentLogMoney 代理端余额日志
type AgentLogMoney struct {
	Common.Common
}

func NewAgentLogMoneyController() *AgentLogMoney {
	return &AgentLogMoney{}
}

// Info 获取详情
func (C *AgentLogMoney) Info(c *gin.Context) {
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

// GetList 代理端余额日志列表（按当前用户过滤）
func (C *AgentLogMoney) GetList(c *gin.Context) {
	var 请求 request.ListLog
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogMoney{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetListByUser(&tx, 请求, c.GetString("User"))
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "获取成功", c)
}

type 请求_AgentLogMoneyBatchDelete struct {
	Id       []int  `json:"Id"`
	Type     int    `json:"Type"`
	Keywords string `json:"Keywords"`
}

// Delete 代理端批量删除（仅能删除自己的日志）
func (C *AgentLogMoney) Delete(c *gin.Context) {
	var 请求 请求_AgentLogMoneyBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogMoney{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDeleteByUser(&tx, 请求.Id, 请求.Type, 请求.Keywords, c.GetString("User"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}
