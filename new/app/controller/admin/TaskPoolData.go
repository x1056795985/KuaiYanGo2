package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	db "server/structs/db"
	"strconv"
)

type TaskPoolData struct {
	Common.Common
}

func NewTaskPoolDataController() *TaskPoolData {
	return &TaskPoolData{}
}

// Index
// @action 任务数据列表
// @show  1
func (C *TaskPoolData) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
		Tid int `json:"Tid"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewTaskPoolData(c, &tx)
	var dataList []db.DB_TaskPoolData
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(请求.List, 请求.Tid, 0, 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "操作成功", c)
	return
}

// Delete
// @action 删除
// @show  2
func (C *TaskPoolData) Delete(c *gin.Context) {
	var 请求 struct {
		UuidS []string `json:"Uuids" binding:"required,min=1"` //id数组
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	tx := *global.GVA_DB
	var S = service.NewTaskPoolData(c, &tx)

	影响行数, err := S.Delete(请求.UuidS)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}



