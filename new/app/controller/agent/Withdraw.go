package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
)

type Withdraw struct {
	Common.Common
}

func NewWithdrawController() *Withdraw {
	return &Withdraw{}
}

func (C *Withdraw) GetConfig(c *gin.Context) {
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	data, err := S.GetAgentConfig(&tx, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}

func (C *Withdraw) UploadPayeeQr(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择收款码图片", c)
		return
	}
	db := *global.GVA_DB
	s := service.S_RmbWithdraw{}
	cfg := s.GetConfig(&db)
	if !cfg.Enable {
		response.FailWithMessage("代理提现未启用", c)
		return
	}
	var S = service.S_RmbWithdraw{}
	path, err := S.UploadPayeeQr(c.GetInt("Uid"), file)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"path": path}, "上传成功", c)
}

func (C *Withdraw) Image(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	info, err := S.GetAgentImage(&tx, c.GetInt("Uid"), req.Path)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.File(info.AbsPath)
}

func (C *Withdraw) Create(c *gin.Context) {
	var req service.WithdrawCreateRequest
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	info, err := S.Create(c.GetInt("Uid"), c.GetString("User"), c.ClientIP(), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(info, "提交成功", c)
}

func (C *Withdraw) List(c *gin.Context) {
	var req service.WithdrawListRequest
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	count, list, err := S.List(&tx, req, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: list, Count: count}, "获取成功", c)
}

func (C *Withdraw) Detail(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	data, err := S.Detail(&tx, req.Id, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}

func (C *Withdraw) Cancel(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	if err := S.Cancel(req.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("取消成功", c)
}
