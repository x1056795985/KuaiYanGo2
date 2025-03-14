package rmbPayItem

import (
	"EFunc/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"net/http"
	"net/url"
	"server/new/app/logic/agent/L_setting"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	utils2 "server/utils"
	"strconv"
	"strings"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_小叮当)
}

var pay_小叮当 小叮当

type 小叮当 struct {
}

func (j 小叮当) Q取通道名称() string {
	return "小叮当"
}

// 当无法通过订单号,获取订单信息时将循环每个接口,尝试获取订单号
func (j 小叮当) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {
	var 局_支付配置 m.Z在线支付_小叮当
	_ = json.Unmarshal(参数.Z支付配置, &局_支付配置)

	//order_no=123456&subject=&pay_type=43&money=10.00&realmoney=10.00&result=success&xddpay_order=654321&app_id=10088&extra=abc
	局_sign := fmt.Sprintf("order_no=%s&subject=%s&pay_type=%s&money=%s&realmoney=%s&result=success&xddpay_order=%s&app_id=%s&extra=%s&",
		c.PostForm("order_no"),
		c.PostForm("subject"),
		c.PostForm("pay_type"),
		c.PostForm("money"),
		c.PostForm("realmoney"),
		c.PostForm("xddpay_order"),
		c.PostForm("app_id"),
		c.PostForm("extra"),
	)

	局_sign = utils2.Md5String(局_sign + 局_支付配置.X小叮当接口密钥)

	if strings.ToUpper(局_sign) != strings.ToUpper(c.PostForm("sign")) {
		return ""
	}
	return c.PostForm("order_no")
}

func (j 小叮当) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_小叮当
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	if err != nil || !局_支付配置.X小叮当支付开关 {
		err = errors.New(局_支付配置.X小叮当支付显示名称 + "支付方式已关闭")
		return
	}
	if 局_支付配置.X小叮当接口密钥 == "" || 局_支付配置.X小叮当app_id == "" {
		err = errors.New(局_支付配置.X小叮当支付显示名称 + "服务端未配置参数")
		return
	}
	if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_积分充值 { //余额充值 和 积分充值判断单次最大金额
		if 参数.Rmb > float64(局_支付配置.X小叮当单次最大金额) {
			err = errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.X小叮当单次最大金额))
			return
		}
	}

	局_网址 := `https://gateway.xddpay.com`
	Http请求 := req.SetRedirectPolicy(req.NoRedirectPolicy()).R()

	//考虑了一下还是项支付宝一样显示url,比较好
	/*	局_网址 += `?format={format}`
		Http请求.SetPathParam(`format`, `json`)*/
	values := url.Values{}
	values.Set("order_no", 参数.PayOrder)
	values.Set("subject", 参数.S商品名称)
	values.Set("pay_type", strconv.Itoa(局_支付配置.X小叮当支付类型))
	values.Set("money", utils.Float64到文本(参数.Rmb, 2))
	values.Set("app_id", 局_支付配置.X小叮当app_id)
	values.Set("extra", "")

	局_sign := utils2.Md5String(fmt.Sprintf("order_no=%s&subject=%s&pay_type=%s&money=%s&app_id=%s&extra=%s&%s",
		values.Get("order_no"),
		values.Get("subject"),
		values.Get("pay_type"),
		values.Get("money"),
		values.Get("app_id"),
		values.Get("extra"),
		局_支付配置.X小叮当接口密钥,
	))

	values.Set("sign", 局_sign)
	post数据 := values.Encode()
	Http请求.SetBodyString(post数据)
	Http请求.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	var 局_请求结果 *req.Response

	for i := 0; i < 3; i++ { // 重试三次防止意外
		局_请求结果, err = Http请求.Post(局_网址)
		if len(局_请求结果.Bytes()) > 0 || err != nil {
			break
		}
	}

	//<html><head><title>Object moved</title></head><body>
	//<h2>Object moved to <a href="/pay/?order_no=ttt123465&amp;subject=YuEr&amp;pay_type=43&amp;money=0.01&amp;app_id=17088&amp;extra=89757&amp;sign=271971B0E9DBFFD85DDC89083FBAB844&amp;xddpay_order=20230907145540699737&amp;user_id=8109&amp;token=3E8F15B80BDC4FC75F798CABD654921B">here</a>.</h2>
	//</body></html>

	jsonStr := strings.Replace(utils.W文本_取出中间文本(局_请求结果.String(), `Object moved to <a href="`, `"`), "&amp;", "&", -1)

	if jsonStr == "" {
		err = errors.New("支付地址获取失败:" + 局_请求结果.String())
		return
	}
	// 处理返回结果

	response = m.Request{
		Status:  1,
		PayURL:  局_网址 + jsonStr,
		OrderId: 参数.PayOrder,
	}
	return
}
func (j 小叮当) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {
	return errors.New("支付类型不支持退款")
}
func (j 小叮当) D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = "err"
			响应代码 = http.StatusInternalServerError
		}
	}()

	var 局_支付配置 m.Z在线支付_小叮当
	if 参数.ReceivedUid == 0 {
		err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	} else {
		局_临时, err2 := L_setting.Q取代理在线支付信息(c, 参数.ReceivedUid)
		if err2 != nil {
			err = errors.Join(errors.New("Q取代理在线支付信息"), err2)
			return
		}
		局_支付配置 = 局_临时.Z在线支付_小叮当
	}

	//order_no=123456&subject=&pay_type=43&money=10.00&realmoney=10.00&result=success&xddpay_order=654321&app_id=10088&extra=abc
	局_sign := fmt.Sprintf("order_no=%s&subject=%s&pay_type=%s&money=%s&realmoney=%s&result=success&xddpay_order=%s&app_id=%s&extra=%s&",
		c.PostForm("order_no"),
		c.PostForm("subject"),
		c.PostForm("pay_type"),
		c.PostForm("money"),
		c.PostForm("realmoney"),
		c.PostForm("xddpay_order"),
		c.PostForm("app_id"),
		c.PostForm("extra"),
	)

	局_sign = utils2.Md5String(局_sign + 局_支付配置.X小叮当接口密钥)

	if strings.ToUpper(局_sign) != strings.ToUpper(c.PostForm("sign")) {
		err = errors.Join(errors.New(c.Request.PostForm.Encode()), errors.New("小叮当异步回调签名验证失败"))
		return
	}

	if 参数.PayOrder == c.PostForm("order_no") && c.PostForm("result") == "success" {
		if c.PostForm("money") != c.PostForm("realmoney") {
			真实金额, err2 := strconv.ParseFloat(c.PostForm("realmoney"), 64)
			if err2 == nil {
				参数.Rmb = 真实金额
			}
		}
		参数.PayOrder2 = c.PostForm("xddpay_order")
	} else {
		err = errors.New(c.PostForm("result"))
	}

	return
}
func (j 小叮当) D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	return
}
