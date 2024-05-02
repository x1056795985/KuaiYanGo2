package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/logic/common/rmbPay"
)

type PayNotify struct {
}

func NewPayNotifyController() *PayNotify {
	return &PayNotify{}
}

// 在线支付通用支付回调
func (s *PayNotify) PayNotify(c *gin.Context) {
	响应信息, 响应代码 := rmbPay.L_rmbPay.D订单回调(c)
	//因为每个平台响应信息都不一样, 所以这个接口,由底层返回响应信息文本和状态码
	c.String(响应代码, 响应信息)
}

// 在线支付通用退款回调
func (s *PayNotify) PayNotify2(c *gin.Context) {
	响应信息, 响应代码 := rmbPay.L_rmbPay.D订单退款回调(c)
	//因为每个平台响应信息都不一样, 所以这个接口,由底层返回响应信息文本和状态码
	c.String(响应代码, 响应信息)
}
