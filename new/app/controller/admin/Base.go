package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"server/Service/Captcha"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	"time"
)

type BaseCtrl struct {
	Common.Common
}

func NewBaseController() *BaseCtrl {
	return &BaseCtrl{}
}

// 验证码api data
type sysCaptchaResponse struct {
	CaptchaId     string `json:"captchaId"`
	PicPath       string `json:"picPath"`
	CaptchaLength int    `json:"captchaLength"`
	OpenCaptcha   bool   `json:"openCaptcha"`
}

// Captcha2 点选验证码
func (b *BaseCtrl) Captcha2(c *gin.Context) {
	openCaptcha := global.GVA_CONFIG.Captcha.OpenCaptcha
	openCaptchaTimeOut := global.GVA_CONFIG.Captcha.OpenCaptchaTimeOut
	key := c.ClientIP()
	v, ok := global.H缓存.Get(key)
	if !ok {
		global.H缓存.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var oc bool
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		oc = true
	}

	验证码id, Base64验证码图片, err := Captcha.Captcha_取点选验证码(interfaceToInt(v))
	if err != nil {
		global.GVA_LOG.Error("验证码获取失败!", zap.Error(err))
		response.FailWithMessage("验证码获取失败", c)
		return
	}

	response.OkWithDetailed(sysCaptchaResponse{
		CaptchaId:     验证码id,
		PicPath:       Base64验证码图片,
		CaptchaLength: global.GVA_CONFIG.Captcha.KeyLong,
		OpenCaptcha:   oc,
	}, "验证码获取成功", c)
}

// 类型转换
func interfaceToInt(v interface{}) (i int) {
	switch v := v.(type) {
	case int:
		i = v
	default:
		i = 0
	}
	return
}
