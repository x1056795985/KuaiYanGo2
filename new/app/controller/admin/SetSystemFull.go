package controller

import (
	"EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/api/middleware"
	"server/config"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/setting"
	m "server/new/app/models/common"
	"server/structs/Http/response"
	utils2 "server/utils"
	"encoding/base64"
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
	var 请求 config.X系统设置
	if !C.ToJSON(c, &请求) {
		return
	}
	err := setting.Z系统设置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	middleware.J集_UserAPi路由_加密.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	response.OkWithMessage("保存成功", c)
}

// S生成API加密源码SDK 生成API加密SDK
func (C *SetSystemFull) S生成API加密源码SDK(c *gin.Context) {
	var 请求 struct {
		Y用户API加密盐 string `json:"用户API加密盐"`
		Type      string `json:"type"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	response.FailWithMessage("生成模块维护中", c)
	return

	if 请求.Y用户API加密盐 == "" {
		response.FailWithMessage("API加密盐值不能为空", c)
		return
	}

	APi列表 := make([]string, 0, len(middleware.J集_UserAPi路由)+1)
	APi列表 = append(APi列表, "GetToken")
	for key := range middleware.J集_UserAPi路由 {
		APi列表 = append(APi列表, key)
	}
	var SDK []byte
	switch 请求.Type {
	case "E":
		SDK = utils2.Y易源码替换APi接口并修复(nil, APi列表, 请求.Y用户API加密盐)
	}
	if len(SDK) == 0 {
		response.FailWithMessage("暂不支持:"+请求.Type, c)
	} else {
		response.OkWithDetailed(
			gin.H{
				"Name":       "飞鸟快验APi加密盐值" + 请求.Y用户API加密盐 + ".e",
				"Base64Data": base64.StdEncoding.EncodeToString(SDK),
			}, "生成成功,记得保存配置使功能生效", c)
	}
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
	var 请求 config.D短信平台配置
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
		err = Captcha.TX云_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 2:
		err = Captcha.D短信宝_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 3:
		err = Captcha.Q七牛云_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 4:
		err = Captcha.K快验_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 5:
		err = Captcha.ALY阿里云_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
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
	var 请求 config.X行为验证码平台配置
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
	var 请求 config.Y云存储配置
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
	var 请求 config.Y用户消息配置
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
	var 请求 config.XAIConfig
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
