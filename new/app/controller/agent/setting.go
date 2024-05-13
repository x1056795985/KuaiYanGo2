package controller

import (
	"github.com/gin-gonic/gin"
	"server/config"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/agent/L_setting"
	"server/new/app/models/db"
	"server/new/app/service"
	"server/structs/Http/response"
	"strings"
)

type Setting struct {
	Common.Common
}

func NewSettingController() *Setting {
	return &Setting{}
}

// 获取代理在线支付信息
func (s *Setting) GetPayInfo(c *gin.Context) {
	data, err := L_setting.Q取代理在线支付信息(c, c.GetInt("Uid"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "操作成功", c)
}

// 置代理在线支付信息
func (s *Setting) SetPayInfo(c *gin.Context) {
	var 请求 config.Z在线支付
	if !s.ToJSON(c, &请求) {
		return
	}
	err := L_setting.Z置代理在线支付信息(c, 请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

type 请求_代理基础设置 struct {
	PromotionCode string `json:"PromotionCode" binding:"required,alphanum,min=1,max=190" zh:"推广码"`
}

// 置代理基础设置
func (s *Setting) SetBaseInfo(c *gin.Context) {
	var 请求 请求_代理基础设置
	if !s.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	_, err := service.NewPromotionCode(c, &tx).Save(db.DB_PromotionCode{c.GetInt("Uid"), 请求.PromotionCode})
	if err != nil {
		局返回 := err.Error()
		if strings.Index(局返回, "Duplicate") != -1 { //唯一索引触发,
			局返回 = "推广码已被其他用户使用,请重新输入"
		}

		response.FailWithMessage(局返回, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// 置代理基础设置
func (s *Setting) GetBaseInfo(c *gin.Context) {
	var 响应 请求_代理基础设置
	tx := *global.GVA_DB
	局_推广信息, err := service.NewPromotionCode(c, &tx).Info(c.GetInt("Uid"))
	if err == nil {
		响应.PromotionCode = 局_推广信息.PromotionCode
	}
	response.OkWithDetailed(响应, "操作成功", c)
}
