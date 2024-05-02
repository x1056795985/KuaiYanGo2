package controller

import (
	"github.com/gin-gonic/gin"
	"server/config"
	"server/new/app/controller/Common"
	"server/new/app/logic/agent/L_setting"
	"server/structs/Http/response"
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
