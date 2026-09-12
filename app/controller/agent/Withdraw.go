package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/global"
	rmbWithdrawLogic "server/app/logic/common/rmbWithdraw"
	"server/app/models/old/response"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
)

type Withdraw struct {
	Common.Common
}

func NewWithdrawController() *Withdraw {
	return &Withdraw{}
}

func (j *Withdraw) GetConfig(c *gin.Context) {
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数据, err := 局_提现服务.Q取代理配置(局_数据库, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(局_数据, "获取成功", c)
}

func (j *Withdraw) UploadPayeeQr(c *gin.Context) {
	局_文件, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择收款码图片", c)
		return
	}
	局_数据库 := global.Get局db()
	局_提现服务 := service.S_RmbWithdraw{}
	局_配置 := 局_提现服务.Q取配置(局_数据库)
	if !局_配置.Enable {
		response.FailWithMessage("代理提现未启用", c)
		return
	}
	局_路径, err := 局_提现服务.S上传收款码(c.GetInt("Uid"), 局_文件)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"path": 局_路径}, "上传成功", c)
}

func (j *Withdraw) Image(c *gin.Context) {
	var 局_请求 struct {
		Path string `json:"path" binding:"required"`
	}
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_图片信息, err := 局_提现服务.Q取代理图片(局_数据库, c.GetInt("Uid"), 局_请求.Path)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.File(局_图片信息.AbsPath)
}

func (j *Withdraw) Create(c *gin.Context) {
	var 局_请求 service.T提现_创建请求
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_信息, err := rmbWithdrawLogic.T提现_创建(global.Get局db(), c.GetInt("Uid"), c.GetString("User"), c.ClientIP(), 局_请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(局_信息, "提交成功", c)
}

func (j *Withdraw) List(c *gin.Context) {
	var 局_请求 service.T提现_列表请求
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数量, 局_列表, err := 局_提现服务.L列表(局_数据库, 局_请求, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: 局_列表, Count: 局_数量}, "获取成功", c)
}

func (j *Withdraw) Detail(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数据, err := 局_提现服务.X详情(局_数据库, 局_请求.Id, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(局_数据, "获取成功", c)
}

func (j *Withdraw) Cancel(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if err := rmbWithdrawLogic.T提现_取消(global.Get局db(), 局_请求.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("取消成功", c)
}
