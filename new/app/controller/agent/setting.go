package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/agent/L_setting"
	m "server/new/app/models/common"
	"server/new/app/models/db"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
	"time"
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
	var 请求 m.Z在线支付
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

// 获取代理云配置
func (s *Setting) GetAgentUserConfig(c *gin.Context) {
	tx := *global.GVA_DB
	var infos []DB.DB_UserConfig
	var err error
	infos, err = service.NewUserConfig(c, &tx).Infos(map[string]interface{}{
		"Uid":   c.GetInt("Uid"),
		"AppId": 50,
	})
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(infos, "操作成功", c)
}

// Del代理云配置
func (s *Setting) DelAgentUserConfig(c *gin.Context) {
	var 请求 struct {
		Name string `json:"Name"  binding:"required,min=1,max=190" zh:"变量名"`
	}
	if !s.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	infos, err := service.NewUserConfig(c, &tx).Delete2(map[string]interface{}{
		"Uid":   c.GetInt("Uid"),
		"AppId": 50,
		"Name":  请求.Name,
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(infos, "操作成功", c)
}

// NewUserConfig信息
func (s *Setting) NewAgentUserConfig(c *gin.Context) {
	var 请求 struct {
		Name string `json:"Name"  binding:"required,min=1,max=190" zh:"变量名"`
	}
	if !s.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB

	_, err := service.NewUserConfig(c, &tx).Info2(map[string]interface{}{"AppId": 50, "Uid": c.GetInt("Uid"), "Name": 请求.Name})

	if err == nil {
		response.FailWithMessage("变量名已存在", c)
		return
	}
	var 代理云配置 DB.DB_UserConfig
	代理云配置.Time = time.Now().Unix()
	代理云配置.UpdateTime = time.Now().Unix()
	代理云配置.User = c.GetString("User")
	代理云配置.Uid = c.GetInt("Uid")
	代理云配置.AppId = 50
	代理云配置.Name = 请求.Name
	代理云配置.Value = ""
	if _, err = service.NewUserConfig(c, &tx).Create(代理云配置); err != nil {
		response.FailWithMessage("添加失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
	return
}

func (s *Setting) SaveAgentUserConfig(c *gin.Context) {

	var 请求 []struct {
		Name  string `json:"Name"  binding:"required,min=1,max=190" zh:"变量名"`
		Value string `json:"Value"   `
	}
	if !s.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB

	for 索引, _ := range 请求 {
		_, err := service.NewUserConfig(c, &tx).Update(map[string]interface{}{
			"AppId": 50,
			"Uid":   c.GetInt("Uid"),
			"Name":  请求[索引].Name,
		}, map[string]interface{}{
			"Value": 请求[索引].Value,
		})
		if err != nil {
			response.FailWithMessage(请求[索引].Name+",保存失败:"+err.Error(), c)
		}
	}

	response.OkWithMessage("保存成功", c)
	return
}
