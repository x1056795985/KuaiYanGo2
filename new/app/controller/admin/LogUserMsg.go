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

// LogUserMsg 用户消息日志
type LogUserMsg struct {
	Common.Common
}

func NewLogUserMsgController() *LogUserMsg {
	return &LogUserMsg{}
}

// Info 获取详情
func (C *LogUserMsg) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

// GetList 用户消息日志列表
func (C *LogUserMsg) GetList(c *gin.Context) {
	var 请求 request.ListLog
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetList(&tx, 请求)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	response.OkWithDetailed(struct {
		List  interface{} `json:"list"`
		Count int64       `json:"count"`
	}{dataList, 总数}, "获取成功", c)
}

type 请求_LogUserMsgBatchDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`     // 1删除ID数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前 7关键字
	Keywords string `json:"keywords"`
}

// Delete 批量删除
func (C *LogUserMsg) Delete(c *gin.Context) {
	var 请求 请求_LogUserMsgBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDelete(&tx, 请求.Id, 请求.Type, 请求.Keywords)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

type 请求_LogUserMsgSetIsRead struct {
	Id     []int `json:"id"`
	IsRead bool  `json:"isRead"`
	Type   int   `json:"type"` // 1修改数组内消息 2全部已读
}

// SetIsRead 批量修改已读状态
func (C *LogUserMsg) SetIsRead(c *gin.Context) {
	var 请求 请求_LogUserMsgSetIsRead
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	err := S.SetIsRead(&tx, 请求.Id, 请求.Type, 请求.IsRead)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// S删除重复消息 删除重复的消息记录
func (C *LogUserMsg) S删除重复消息(c *gin.Context) {
	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	err := S.S删除重复消息(&tx)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}
