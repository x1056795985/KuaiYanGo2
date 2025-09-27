package rmbPayItem

import (
	"EFunc/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/imroc/req/v3"
	"net/http"
	"net/url"
	"server/new/app/logic/agent/L_setting"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	utils2 "server/utils"
	"sort"
	"strconv"
	"strings"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_易支付2)
}

var pay_易支付2 易支付2

type 易支付2 struct {
}

func (j 易支付2) Q取通道名称() string {
	return "易支付2"
}

// 当无法通过订单号,获取订单信息时将循环每个接口,尝试获取订单号
func (j 易支付2) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {
	var 局_支付配置 m.Z在线支付_易支付2
	_ = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	//
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

	局_sign = utils2.Md5String(局_sign + 局_支付配置.Y易支付2商户密钥KEY)

	if strings.ToUpper(局_sign) != strings.ToUpper(c.PostForm("sign")) {
		return ""
	}
	return c.PostForm("order_no")
}

func (j 易支付2) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_易支付2
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	if err != nil || !局_支付配置.Y易支付2开关 {
		err = errors.New(局_支付配置.Y易支付2显示名称 + "支付方式已关闭")
		return
	}
	if 局_支付配置.Y易支付2网关 == "" || 局_支付配置.Y易支付2商户密钥KEY == "" || 局_支付配置.Y易支付2商户ID == "" {
		err = errors.New(局_支付配置.Y易支付2显示名称 + "服务端未配置参数")
		return
	}

	if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_积分充值 { //余额充值 和 积分充值判断单次最大金额
		if 参数.Rmb > float64(局_支付配置.Y易支付2最大金额) {
			err = errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Y易支付2最大金额))
			return
		}
	}
	//http://127.0.0.1:18888/Admin   获取http://127.0.0.1:18888 部分
	// 解析网关URL获取基础地址
	parsedURL, err := url.Parse(局_支付配置.Y易支付2网关)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		err = errors.New("网关地址格式错误")
		return
	}
	// 组合协议和主机地址
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host
	if parsedURL.Path == "" {
		parsedURL.Path = "/submit.php"
	}
	局_网址 := baseURL + parsedURL.Path //易支付2v1通用

	Http请求 := req.SetRedirectPolicy(req.NoRedirectPolicy()).R()
	values := url.Values{}
	values.Set("money", utils.Float64到文本(参数.Rmb, 2))
	values.Set("name", 参数.S商品名称)
	values.Set("notify_url", 参数.Y异步回调地址)
	values.Set("out_trade_no", 参数.PayOrder)
	values.Set("pid", 局_支付配置.Y易支付2商户ID)
	values.Set("return_url", rmbPay.L_rmbPay.Z支付订单回调关键字转换(局_支付配置.Y易支付2同步回调url, 参数))
	values.Set("type", 局_支付配置.Y易支付2支付方式)
	values.Set("sitename", 参数.S商品名称)
	values.Set("clientip", c.ClientIP())

	// 过滤掉不需要参与签名的参数
	var 参数列表 []string
	for key, value := range values {
		if key != "sign" && key != "sign_type" && value[0] != "" {
			参数列表 = append(参数列表, key)
		}
	}

	// 按参数名ASCII码排序
	sort.Strings(参数列表)

	// 拼接参数
	var 拼接字符串 strings.Builder
	for i, key := range 参数列表 {
		拼接字符串.WriteString(key)
		拼接字符串.WriteString("=")
		拼接字符串.WriteString(values.Get(key))
		if i < len(参数列表)-1 {
			拼接字符串.WriteString("&")
		}
	}

	// 生成签名
	局_sign := utils2.Md5String(拼接字符串.String() + 局_支付配置.Y易支付2商户密钥KEY)

	values.Set("sign", 局_sign)
	values.Set("sign_type", "MD5")
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
	if err != nil {
		err = errors.New("支付地址请求失败:" + err.Error())
		return
	}
	//<script>window.location.href='/Pay/console?trade_no=Y2025032602391019226';</script>
	//<script>window.location.href='/Pay/console?trade_no=Y2025032602472053282';</script>
	//{"code":1,"msg":"","trade_no":"20250326195924890686","payurl":"","qrcode":"https://yz.mmlwo.cn/api/pay/toapp/20250326195924890686","urlscheme":"alipays://platformapi/startapp?appId=20000067\u0026url=https%3A%2F%2Frender.alipay.com%2Fp%2Fs%2Fi%3Fscheme%3Dalipays%253A%252F%252Fplatformapi%252Fstartapp%253FappId%253D20000180%2526url%253Dhttps%25253A%25252F%25252Fyz.mmlwo.cn%25252Fapi%25252Fpay%25252Ftoapp%25252F20250326195924890686","money":"0.02"}
	//{"success":true,"msg":"","code":1,"trade_no":"499059861254979584","payurl":"https://alipaypage3glj1qtw0xz4.zhifu.fm.it88168.com/pay?orderNo=499059861254979584","qrcode":null,"extend_params":null}
	//{"code":1,"msg":"success","trade_no":"20250926214316327191","qrcode":"https://render.alipay.com/p/s/i?scheme=alipays%3A%2F%2Fplatformapi%2Fstartapp%3FappId%3D20000116%26actionType%3DtoAccount%26goBack%3DNO%26amount%3D0.03%26userId%3D2088222179021701%26memo%3D20250926214316327191","urlscheme":"alipayqr://platformapi/startapp?appId=20000067\u0026url=https%3A%2F%2Frender.alipay.com%2Fp%2Fs%2Fi%3Fscheme%3Dalipays%253A%252F%252Fplatformapi%252Fstartapp%253FappId%253D20000116%2526actionType%253DtoAccount%2526goBack%253DNO%2526amount%253D0.03%2526userId%253D2088222179021701%2526memo%253D20250926214316327191","money":"0.03"}
	//判断是否为json
	if strings.HasPrefix(局_请求结果.String(), "{") {
		局_json := gjson.New(局_请求结果.String())
		if 局_json.Get("code").Int() != 1 {
			err = errors.New("支付地址获取失败:" + 局_请求结果.String())
			return
		}

		if 局_json.Get("qrcode").String() != "" {
			response = m.Request{
				Status:       1,
				PayURL:       局_json.Get("qrcode").String(),
				OrderId:      参数.PayOrder,
				PayQRCode:    局_json.Get("urlscheme").String(),
				PayQRCodePNG: rmbPay.L_rmbPay.S生成二维码并转base64(局_json.Get("urlscheme").String()),
			}
		} else if 局_json.Get("payurl").String() != "" {
			response = m.Request{
				Status:       1,
				PayURL:       局_json.Get("payurl").String(),
				OrderId:      参数.PayOrder,
				PayQRCode:    局_json.Get("qrcode").String(),
				PayQRCodePNG: rmbPay.L_rmbPay.S生成二维码并转base64(局_json.Get("qrcode").String()),
			}
		} else { //二开后的格式太多了,返回给客户自己处理吧
			response = m.Request{
				Status:       1,
				PayURL:       "",
				OrderId:      参数.PayOrder,
				PayQRCode:    "",
				PayQRCodePNG: "",
				Other:        局_请求结果.String(),
			}
		}
		//判断response.PayQRCode  左边是否为http  如果不是http,就改生成 PayURL值的二维码
		if !strings.HasPrefix(response.PayQRCode, "http") {
			response.PayQRCodePNG = rmbPay.L_rmbPay.S生成二维码并转base64(response.PayURL)
		}

		return
	}

	jsonStr := strings.Replace(utils.W文本_取出中间文本(局_请求结果.String(), `href='`, `'`), "&amp;", "&", -1)
	if jsonStr == "" {
		err = errors.New("支付地址获取失败:" + 局_请求结果.String())
		return
	}
	// 处理返回结果

	if utils.W文本_取左边(jsonStr, 4) == "http" {
		局_网址 = jsonStr
	} else {
		局_网址 = baseURL + jsonStr
	}

	response = m.Request{
		Status:       1,
		PayURL:       局_网址,
		OrderId:      参数.PayOrder,
		PayQRCode:    局_网址,
		PayQRCodePNG: rmbPay.L_rmbPay.S生成二维码并转base64(局_网址),
	}
	return
}
func (j 易支付2) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {
	return errors.New("支付类型不支持退款")
}
func (j 易支付2) D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = "fail" // 与PHP保持一致的返回标识
			响应代码 = http.StatusInternalServerError
		}
	}()

	var 局_支付配置 m.Z在线支付_易支付2
	if 参数.ReceivedUid == 0 {
		err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	} else {
		局_临时, err2 := L_setting.Q取代理在线支付信息(c, 参数.ReceivedUid)
		if err2 != nil {
			err = errors.Join(errors.New("Q取代理在线支付信息"), err2)
			return
		}
		局_支付配置 = 局_临时.Z在线支付_易支付2
	}

	// 参数获取方式改为GET（根据易支付2文档确认）
	请求参数 := c.Request.URL.Query()
	//pid=1002&trade_no=Y2025032610445366076&out_trade_no=250326104453000001&type=alipay&name=%E6%94%AF%E4%BB%98%E8%B4%AD%E5%8D%A1%3A_%E6%94%AF%E4%BB%98%E8%B4%AD%E5%8D%A1&money=0.02&trade_status=TRADE_SUCCESS&sign=5ec6401b4bb0b14bfc308b5e56d58b7b&sign_type=MD5
	// 修改后的签名验证逻辑（包含所有参数）
	参数列表 := make([]string, 0)
	for key := range 请求参数 {
		// 排除签名相关参数
		if key != "sign" && key != "sign_type" && 请求参数.Get(key) != "" {
			参数列表 = append(参数列表, key)
		}
	}

	// 按参数名ASCII码排序
	sort.Strings(参数列表)

	// 按参数名ASCII码排序
	sort.Strings(参数列表)

	// 拼接参数
	var 签名参数 strings.Builder
	for i, key := range 参数列表 {
		值 := 请求参数.Get(key)
		签名参数.WriteString(fmt.Sprintf("%s=%s", key, 值))
		if i < len(参数列表)-1 {
			签名参数.WriteString("&")
		}
	}
	签名参数.WriteString(局_支付配置.Y易支付2商户密钥KEY) // 注意易支付2密钥拼接方式

	局_sign := strings.ToUpper(utils2.Md5String(签名参数.String()))

	if 局_sign != strings.ToUpper(请求参数.Get("sign")) {
		err = errors.New("易支付2异步回调签名验证失败")
		return
	}

	// 严格匹配交易状态
	if 请求参数.Get("trade_status") != "TRADE_SUCCESS" {
		err = errors.New("交易未成功:" + 请求参数.Get("trade_status"))
		return
	}

	// 验证订单号一致性
	if 参数.PayOrder != 请求参数.Get("out_trade_no") {
		err = errors.New("订单号不匹配")
		return
	}

	// 记录易支付2交易号
	参数.PayOrder2 = 请求参数.Get("trade_no")

	// 处理金额转换
	if 金额, err2 := strconv.ParseFloat(请求参数.Get("money"), 64); err2 == nil {
		参数.Rmb = 金额
	} else {
		err = errors.Join(errors.New("金额解析错误"), err2)
	}

	return
}

func (j 易支付2) D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	return
}
