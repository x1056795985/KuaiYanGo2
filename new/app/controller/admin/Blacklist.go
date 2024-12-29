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

// Blacklist
// @MenuName 日志管理
// @ModuleName 黑名单
type Blacklist struct {
	Common.Common
}

func NewBlacklistController() *Blacklist {
	var C = Blacklist{}
	return &C
}

type 请求_Create struct {
	AppId   int    `json:"AppId" binding:"required"`
	ItemKey string `json:"ItemKey" binding:"required,min=1,max=190" zh:"拉黑信息"` // 索引最大长度767字节 除4 就是191  否则INNODB引擎报错  Specified key wastoo long; max key length is 767 bytes
	Note    string `json:"Note" binding:"max=1000" zh:"备注"`
}

// Create
// @action 添加
// @show  2
func (C *Blacklist) Create(c *gin.Context) {
	var 请求 请求_Create
	if !C.ToJSON(c, &请求) {
		return
	}
	var S = service.S_Blacklist{}
	tx := *global.GVA_DB
	err := S.Create(&tx, db.DB_Blacklist{AppId: 请求.AppId, ItemKey: 请求.ItemKey, Note: 请求.Note})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.Ok(c)
}

// Delete
// @action 删除
// @show  2
func (C *Blacklist) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_Blacklist{}
	tx := *global.GVA_DB

	影响行数, err := S.Delete(&tx, 请求.Ids)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// Update
// @action 更新
// @show  2
func (C *Blacklist) Update(c *gin.Context) {
	var 请求 db.DB_Blacklist
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id必须大于0", c)
		return
	}

	var S = service.S_Blacklist{}
	tx := *global.GVA_DB
	err := S.Update(&tx, 请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}

	response.OkWithMessage("操作成功", c)
	return
}

// Info
// @action 查询
// @show  2
func (C *Blacklist) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_Blacklist{}
	tx := *global.GVA_DB
	var info db.DB_Blacklist
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(info, "操作成功", c)
	return
}

type 结构请求_GetList struct {
	request.List
	AppId int `json:"AppId"`
}

// Index
// @action 黑名单列表
// @show  1
func (C *Blacklist) GetList(c *gin.Context) {
	var 请求 结构请求_GetList
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_Blacklist{}
	tx := *global.GVA_DB
	var dataList []db.DB_Blacklist
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(&tx, 请求.List, 请求.AppId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList{List: dataList, Count: 总数}, "操作成功", c)
	return
	//继续对接前端
}

type 请求_批量删除 struct {
	Type int `json:"Type" binding:"required,min=1"`
}

// DeleteBatch
// @action 删除批量维护
// @show  2
func (C *Blacklist) DeleteBatch(c *gin.Context) {
	var 请求 请求_批量删除
	if !C.ToJSON(c, &请求) {
		return
	}

	var 影响行数 int64
	var S = service.S_Blacklist{}
	tx := *global.GVA_DB

	影响行数, err := S.DeleteType(&tx, 请求.Type)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}
