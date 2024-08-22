package SetSystem

import (
	"EFunc/utils"
	_ "embed"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/api/middleware"
	"server/config"
	"server/new/app/logic/common/mqttClient"
	"server/new/app/logic/common/setting"
	"server/structs/Http/response"
	utils2 "server/utils"
)

type Api struct{}

func (a *Api) GetInfoSystem(c *gin.Context) {
	//方便手动修改配置后重新读取
	response.OkWithDetailed(setting.Q系统设置(), "获取成功", c)
	return
}

// 暂时放弃  go:embed  \..\..\..\SDK/易语言/飞鸟快验网络验证对接模块.e
var 快验网络验证对接易模块 []byte

type 请求_S生成API加密源码SDK struct {
	Y用户API加密盐 string `mapstructure:"用户API加密盐" json:"用户API加密盐" yaml:"用户API加密盐"`
	Type      string `mapstructure:"Type" json:"Type" yaml:"Type"` //"E"  易源码
}
type 响应_S生成API加密源码SDK struct {
	Name       string `mapstructure:"Name" json:"Name" yaml:"Name"`
	Base64Data string `mapstructure:"Base64Data" json:"Base64Data" yaml:"Base64Data"` //"E"  易源码
}

func (a *Api) S生成API加密源码SDK(c *gin.Context) {
	var 请求 请求_S生成API加密源码SDK
	response.FailWithMessage("生成模块维护中", c)
	return

	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Y用户API加密盐 == "" {
		response.FailWithMessage("API加密盐值不能为空", c)
		return
	}

	// 遍历map获取所有key
	APi列表 := make([]string, 0, len(middleware.J集_UserAPi路由)+1)
	APi列表 = append(APi列表, "GetToken")
	for key := range middleware.J集_UserAPi路由 {
		APi列表 = append(APi列表, key)
	}
	var SDK []byte
	switch 请求.Type {
	case "E":
		SDK = utils2.Y易源码替换APi接口并修复(快验网络验证对接易模块, APi列表, 请求.Y用户API加密盐)
	}
	if len(SDK) == 0 {
		response.FailWithMessage("暂不支持:"+请求.Type, c)
	} else {
		response.OkWithDetailed(
			响应_S生成API加密源码SDK{
				Name:       "AI矩阵兽APi加密盐值" + 请求.Y用户API加密盐 + ".e",
				Base64Data: base64.StdEncoding.EncodeToString(SDK),
			}, "生成成功,记得保存配置使功能生效", c)
	}
	return
}

// save 保存
func (a *Api) Save信息System(c *gin.Context) {
	var 请求 config.X系统设置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	err = setting.Z系统设置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	middleware.J集_UserAPi路由_加密.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	response.OkWithMessage("保存成功", c)
	return
}

func (a *Api) GetInfo在线支付(c *gin.Context) {
	response.OkWithDetailed(setting.Q在线支付配置(), "获取成功", c)
	return
}

// save 保存
func (a *Api) Save信息在线支付(c *gin.Context) {
	var 请求 config.Z在线支付
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	err = setting.Z在线支付配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("保存成功", c)
	return
}

func (a *Api) GetInfo短信平台设置(c *gin.Context) {
	response.OkWithDetailed(setting.Q短信平台配置(), "获取成功", c)
	return
}
func (a *Api) GetInfo行为验证码平台设置(c *gin.Context) {
	response.OkWithDetailed(setting.Q行为验证码平台配置(), "获取成功", c)
	return
}

func (a *Api) GetInfoMQTT配置(c *gin.Context) {
	var 配置 = setting.Q取MQTT配置()
	配置.L连接状态 = mqttClient.L_mqttClient.Q取连接状态(c)
	response.OkWithDetailed(配置, "获取成功", c)
	return
}

func (a *Api) Mqtt发送测试(c *gin.Context) {
	var 请求 struct {
		Z主题 string `json:"主题"`
		N内容 string `json:"内容"`
	}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if !mqttClient.L_mqttClient.Q取连接状态(c) {
		response.FailWithMessage("请先打开连接状态,并保存", c)
		return
	}
	err = mqttClient.L_mqttClient.F发送消息(c, 请求.Z主题, 请求.N内容)
	if err == nil {
		response.OkWithMessage("发送成功", c)
	} else {
		response.FailWithMessage(err.Error(), c)
	}
	return
}

// save 保存
func (a *Api) SaveMQTT配置(c *gin.Context) {
	var 请求 config.MQTT配置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	mqttClient.L_mqttClient.D断开(nil)
	if 请求.L连接状态 {
		err = mqttClient.L_mqttClient.L连接(nil, 请求.F服务器地址, 请求.F服务器端口, 请求.Y用户名, 请求.M密码)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}

	err = setting.Z置MQTT配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

// save 保存
func (a *Api) Save短信平台设置(c *gin.Context) {
	var 请求 config.D短信平台配置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	err = setting.Z短信平台配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

type Save短信平台测试 struct {
	Id    int    `json:"Id"`
	Phone string `json:"Phone"`
}

func (a *Api) F发送短信平台测试(c *gin.Context) {
	var 请求 Save短信平台测试
	err := c.ShouldBindJSON(&请求)
	//解析失败
	switch 请求.Id {
	case 0, 1:
		err = Captcha.TX云_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 2:
		err = Captcha.D短信宝_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 3:
		err = Captcha.Q七牛云_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	case 4:
		err = Captcha.K快验_sms发送短信验证码([]string{utils.W文本_取随机字符串_数字(6)}, 请求.Phone)
	default:
		err = errors.New("短信平台配置.当前选择配置无效")
	}
	if err == nil {
		response.OkWithMessage("测试验证码短信发送成功", c)
	} else {
		response.FailWithMessage(err.Error(), c)
	}
	return
}

// save 保存
func (a *Api) Save行为验证码平台设置(c *gin.Context) {
	var 请求 config.X行为验证码平台配置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	err = setting.Z行为验证码平台配置(&请求)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}
