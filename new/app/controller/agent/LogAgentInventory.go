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

// AgentLogAgentInventory 代理端库存日志
type AgentLogAgentInventory struct {
	Common.Common
}

func NewAgentLogAgentInventoryController() *AgentLogAgentInventory {
	return &AgentLogAgentInventory{}
}

// Info 获取详情
func (C *AgentLogAgentInventory) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentInventory{}
	tx := *global.GVA_DB
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

// GetList 代理端库存日志列表（按User1 OR User2过滤）
func (C *AgentLogAgentInventory) GetList(c *gin.Context) {
	var 请求 request.ListLog
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentInventory{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetListByUser(&tx, 请求, c.GetString("User"))
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "获取成功", c)
}

type 请求_AgentLogAgentInventoryBatchDelete struct {
	Id       []int  `json:"Id"`
	Type     int    `json:"Type"`
	Keywords string `json:"Keywords"`
}

// Delete 代理端批量删除（仅能删除自己相关的日志）
func (C *AgentLogAgentInventory) Delete(c *gin.Context) {
	var 请求 请求_AgentLogAgentInventoryBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentInventory{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDeleteByUser(&tx, 请求.Id, 请求.Type, 请求.Keywords, c.GetString("User"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}
