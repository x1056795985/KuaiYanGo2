package controller

import (
	"EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/logic/common/captcha"
	"server/app/logic/common/setting"
	m "server/app/models/common"
	"server/app/models/old/response"
	"server/app/router/userSafetyApi"
)

type SetSystemFull struct {
	Common.Common
}

func NewSetSystemFullController() *SetSystemFull {
	return &SetSystemFull{}
}

// GetInfoSystem 获取系统设置
func (C *SetSystemFull) GetInfoSystem(c *gin.Context) {
	response.OkWithDetailed(setting.Q系统设置(), "获取成功", c)
}

// SaveInfoSystem 保存系统设置
func (C *SetSystemFull) SaveInfoSystem(c *gin.Context) {
	var 请求 m.X系统设置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z系统设置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	userSafetyApi.J集_UserAPi路由_加密.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	response.OkWithMessage("保存成功", c)
}

// GetInfo在线支付 获取在线支付配置
func (C *SetSystemFull) GetInfo在线支付(c *gin.Context) {
	response.OkWithDetailed(setting.Q在线支付配置(), "获取成功", c)
}

// SaveInfo在线支付 保存在线支付配置
func (C *SetSystemFull) SaveInfo在线支付(c *gin.Context) {
	var 请求 m.Z在线支付
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z在线支付配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// GetInfo短信平台设置 获取短信平台配置
func (C *SetSystemFull) GetInfo短信平台设置(c *gin.Context) {
	response.OkWithDetailed(setting.Q短信平台配置(), "获取成功", c)
}

// Save短信平台设置 保存短信平台配置
func (C *SetSystemFull) Save短信平台设置(c *gin.Context) {
	var 请求 m.D短信平台配置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z短信平台配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// F发送短信平台测试 发送短信测试
func (C *SetSystemFull) F发送短信平台测试(c *gin.Context) {
	var 请求 struct {
		Id    int    `json:"id"`
		Phone string `json:"phone"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	switch 请求.Id {
	case 0, 1:
		err = captcha.SendTencentSMS([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 2:
		err = captcha.SendSMSBao([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 3:
		err = captcha.SendQiniuSMS([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 4:
		err = captcha.SendKuaiYanSMS([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 5:
		err = captcha.SendAliyunSMS([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	default:
		err = errors.New("短信平台配置.当前选择配置无效")
	}
	if err == nil {
		response.OkWithMessage("测试验证码短信发送成功", c)
	} else {
		response.FailWithMessage(err.Error(), c)
	}
}

// GetInfo行为验证码平台设置 获取行为验证码配置
func (C *SetSystemFull) GetInfo行为验证码平台设置(c *gin.Context) {
	response.OkWithDetailed(setting.Q行为验证码平台配置(), "获取成功", c)
}

// Save行为验证码平台设置 保存行为验证码配置
func (C *SetSystemFull) Save行为验证码平台设置(c *gin.Context) {
	var 请求 m.X行为验证码平台配置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z行为验证码平台配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// GetInfo云存储设置 获取云存储配置
func (C *SetSystemFull) GetInfo云存储设置(c *gin.Context) {
	response.OkWithDetailed(setting.Q云存储配置(), "获取成功", c)
}

// Save云存储设置 保存云存储配置
func (C *SetSystemFull) Save云存储设置(c *gin.Context) {
	var 请求 m.Y云存储配置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z云存储配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// Get用户消息配置 获取用户消息配置
func (C *SetSystemFull) Get用户消息配置(c *gin.Context) {
	response.OkWithDetailed(setting.Q用户消息配置(), "获取成功", c)
}

// Save用户消息配置 保存用户消息配置
func (C *SetSystemFull) Save用户消息配置(c *gin.Context) {
	var 请求 m.Y用户消息配置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z用户消息配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// GetInfoAiConfig 获取AI配置
func (C *SetSystemFull) GetInfoAiConfig(c *gin.Context) {
	response.OkWithDetailed(setting.QAI配置(), "获取成功", c)
}

// SaveInfoAiConfig 保存AI配置
func (C *SetSystemFull) SaveInfoAiConfig(c *gin.Context) {
	var 请求 m.XAIConfig
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.ZAI配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}
