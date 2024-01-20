package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"server/global"
	"server/new/app/logic/common/cron/functions"
	"server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/new/app/utils"
	"server/structs/Http/response"
	"strconv"
	"time"
)

// Cron
// @MenuName 二开扩展
// @ModuleName 定时任务
type Cron struct {
}

func NewCronController() *Cron {
	var C = Cron{}
	return &C
}

// 统一反序列化参数
func (C *Cron) ToJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		// 获取validator.ValidationErrors类型的errors
		errs, ok := err.(validator.ValidationErrors)
		errStr := ""
		if !ok {
			errStr = "参数错误:" + err.Error() //	// 非validator.ValidationErrors类型错误直接返回
		} else {
			for _, v := range errs.Translate(global.Trans) { // validator.ValidationErrors类型错误则进行翻译
				errStr += v + ","
			}
		}
		response.FailWithMessage(errStr, c)
		return false
	}
	return true
}

// Create
// @action 添加
// @show  2
func (C *Cron) Create(c *gin.Context) {

	var 请求 struct {
		Name    string `json:"Name" binding:"required"`
		Status  int    `json:"Status" binding:"required"`
		Cron    string `json:"Cron" binding:"required"`
		Type    int    `json:"Type" binding:"required" zh:"类型"`
		RunText string `json:"RunText" binding:"required,min=1,max=1000" zh:"运行数据"`
		Note    string `json:"Note" binding:"max=1000" zh:"备注"`
	}

	if !C.ToJSON(c, &请求) {
		return
	}

	if !utils.IsCron表达式(请求.Cron) {
		response.FailWithMessage("cron表达式不争取,标准6位,秒,分,时,天,月,周", c)
		return
	}
	var S = service.S_Cron{}
	tx := *global.GVA_DB
	err := S.Create(&tx, db.DB_Cron{Name: 请求.Name, Status: 请求.Status, Type: 请求.Type, Cron: 请求.Cron, RunText: 请求.RunText, Note: 请求.Note})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	if 请求.Status == 1 {
		_ = functions.S刷新数据库定时任务(true)
	}

	response.Ok(c)
}

// Delete
// @action 删除
// @show  2
func (C *Cron) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_Cron{}
	tx := *global.GVA_DB

	影响行数, err := S.Delete(&tx, 请求.Ids)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = functions.S刷新数据库定时任务(true)
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// Update
// @action 更新
// @show  2
func (C *Cron) Update(c *gin.Context) {
	var 请求 db.DB_Cron
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id必须大于0", c)
		return
	}
	if !utils.IsCron表达式(请求.Cron) {
		response.FailWithMessage("cron表达式不争取,标准6位,秒,分,时,天,月,周", c)
		return
	}

	var S = service.S_Cron{}
	tx := *global.GVA_DB
	err := S.Update(&tx, 请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	_ = functions.S刷新数据库定时任务(true)
	response.OkWithMessage("操作成功", c)
	return
}

// Info
// @action 查询
// @show  2
func (C *Cron) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_Cron{}
	tx := *global.GVA_DB
	var info db.DB_Cron
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(info, "操作成功", c)
	return
}

// Index
// @action 定时任务列表
// @show  1
func (C *Cron) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		AppId int `json:"AppId"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_Cron{}
	tx := *global.GVA_DB
	var dataList []db.DB_Cron
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(&tx, 请求.List, 请求.AppId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "操作成功", c)
	return
	//继续对接前端
}

// DeleteBatch
// @action 删除批量维护
// @show  2
func (C *Cron) DeleteBatch(c *gin.Context) {
	var 请求 struct {
		Type int `json:"Type" binding:"required,min=1"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_Cron{}
	tx := *global.GVA_DB

	影响行数, err := S.DeleteType(&tx, 请求.Type)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = functions.S刷新数据库定时任务(true)
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// UpdateStatus
// @action 更新状态
// @show  2
func (C *Cron) UpdateStatus(c *gin.Context) {
	var 请求 struct {
		Id     int `json:"Id" binding:"required"`
		Status int `json:"Status" binding:"required" zh:"状态"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id必须大于0", c)
		return
	}

	var S = service.S_Cron{}
	tx := *global.GVA_DB
	CronInfo, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	CronInfo.Status = 请求.Status
	err = S.Update(&tx, CronInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	_ = functions.S刷新数据库定时任务(true)
	response.OkWithMessage("操作成功", c)
	return
}

// Z执行
// @action 执行一次
// @show  2
func (C *Cron) Z执行(c *gin.Context) {
	var 请求 request.Id
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id必须大于0", c)
		return
	}

	var S = service.S_Cron{}
	tx := *global.GVA_DB
	CronInfo, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	通用任务执行函数2, err := functions.T通用任务执行函数2(time.Now().Unix(), CronInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithMessage(通用任务执行函数2, c)
	return
}
