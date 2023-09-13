package base

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"server/Service/Captcha"
	"server/global"
	"server/structs/Http/response"
	"time"
)

// Captcha
// @Tags      Base
// @Summary   生成验证码
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=systemRes.SysCaptchaResponse,msg=string}  "生成验证码,返回包括随机数id,base64,验证码长度,是否开启验证码"
// @Router    /base/captcha [post]
func (b *BaseApi) Captcha(c *gin.Context) {
	// 判断验证码是否开启
	openCaptcha := global.GVA_CONFIG.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.GVA_CONFIG.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	key := c.ClientIP()                                                //获取客户端ip
	v, ok := global.H缓存.Get(key)                                       //获取
	if !ok {
		global.H缓存.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var oc bool
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		oc = true
	}

	验证码id, Base64验证码图片, err := Captcha.Captcha_取英数验证码()
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

// 验证码api  data
type sysCaptchaResponse struct {
	CaptchaId     string `json:"captchaId"`     //验证码id
	PicPath       string `json:"picPath"`       //验证码数据
	CaptchaLength int    `json:"captchaLength"` //验证码长度
	OpenCaptcha   bool   `json:"openCaptcha"`   //是否显示验证码
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
