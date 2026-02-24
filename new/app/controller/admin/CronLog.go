package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

// CronLog
// @MenuName 日志管理
// @ModuleName 定时任务
type CronLog struct {
	Common.Common
}

func NewCronLogController() *CronLog {
	var C = CronLog{}
	return &C
}

// Delete
// @action 删除
// @show  2
func (C *CronLog) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_CronLog{}
	tx := *global.GVA_DB

	影响行数, err := S.Delete(&tx, 请求.Ids)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// Info
// @action 查询
// @show  2
func (C *CronLog) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_CronLog{}
	tx := *global.GVA_DB
	var info db.DB_Cron_log
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(info, "操作成功", c)
	return
}

// Index
// @action 黑名单列表
// @show  1
func (C *CronLog) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		TaskType     int      `json:"TaskType"`
		Result       int8     `json:"Result"`
		RegisterTime []string `json:"RegisterTime"` // 开始时间 结束时间
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_CronLog{}
	tx := *global.GVA_DB
	var dataList []db.DB_Cron_log
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(&tx, 请求.List, 请求.Result, 请求.TaskType, 请求.RegisterTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "操作成功", c)
	return
	//继续对接前端
}

// DeleteBatch
// @action 删除批量维护
// @show  2
func (C *CronLog) DeleteBatch(c *gin.Context) {
	var 请求 struct {
		Type     int    `json:"Type" binding:"required,min=1"`
		Keywords string `json:"Keywords" `
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_CronLog{}
	tx := *global.GVA_DB
	影响行数, err := S.DeleteType(&tx, 请求.Type, 请求.Keywords)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}
