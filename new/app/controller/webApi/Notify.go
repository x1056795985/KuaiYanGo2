package controller

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/Service/Ser_Log"
	"server/Service/Ser_RMBPayOrder"
	"server/api/WebApi"
	"server/new/app/logic/common/setting"
	"sort"
	"strings"
)

type PayNotify struct {
}

func NewPayNotifyController() *PayNotify {
	return &PayNotify{}
}

// 虎皮椒异步回调  Notify - 支付成功后会回调这里;我们可以用来修改订单状态等等
func (s *PayNotify) PayHuPiJiaoNotify(c *gin.Context) {

	/*
	   /**
	    * 回调数据
	    * @var array(
	    *       'trade_order_id'，商户网站订单ID
	            'total_fee',订单支付金额
	            'transaction_id',//支付平台订单ID
	            'order_date',//支付时间
	            'plugins',//自定义插件ID,与支付请求时一致
	            'status'=>'OD'//订单状态，OD已支付，WP未支付
	    *   )
	*/

	data, err := ioutil.ReadAll(c.Request.Body)
	//trade_order_id=202401052126060001&total_fee=0.01&transaction_id=4200001926202401054012972507&open_order_id=20241873681&order_title=%E7%94%A8%E6%88%B7%3Aaaaaaa_%E8%B4%AD%E5%8D%A1%E7%9B%B4%E5%86%B2&status=OD&nonce_str=4217541016&time=1704461215&appid=201906157675&hash=9bdfdb7f14a8bf0db85139a17e6b9372
	//fmt.Printf(string(data))
	if err != nil {
		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "虎皮椒异步回调被异常调用:"+c.Request.RequestURI+"|"+string(data))
		return
	}

	formData, err := url.ParseQuery(string(data))
	if err != nil {
		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "虎皮椒异步回调反序列化失败:"+c.Request.RequestURI+"|"+string(data))
		return
	}
	callbackData := make(map[string]string)
	for key, value := range formData {
		callbackData[key] = value[0]
	}

	局_支付配置 := setting.Q在线支付配置()
	// 验证签名
	appSecret := 局_支付配置.H虎皮椒appSecret
	hash := SignHuPiJiao(callbackData, appSecret)
	if callbackData["hash"] != hash {
		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "虎皮椒异步回调异常调用签名错误:"+c.Request.RequestURI+"|"+string(data))
		return
	}

	// 商户订单ID
	tradeOrderID, ok2 := callbackData["trade_order_id"]
	if !ok2 {
		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "虎皮椒异步回调订单参数错误:"+c.Request.RequestURI+"|"+string(data))
		return
	}

	if callbackData["status"] == "OD" {
		//这里是支付成功的回调
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(tradeOrderID)
		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付
			tradeOrderID, ok = callbackData["transaction_id"]
			if ok {
				局_订单详细信息.PayOrder2 = tradeOrderID
			}
			WebApi.Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
	return
}

// 虎皮椒签名计算

func SignHuPiJiao(params map[string]string, appSecret string) string {
	var data string
	keys := make([]string, 0, 0)
	for key, _ := range params {
		if key != "hash" {
			keys = append(keys, key)
		}

	}
	sort.Strings(keys)
	//拼接
	for _, k := range keys {
		data = fmt.Sprintf("%s%s=%s&", data, k, params[k])
	}
	data = strings.Trim(data, "&")
	data = fmt.Sprintf("%s%s", data, appSecret)
	m := md5.New()
	m.Write([]byte(data))
	sign := fmt.Sprintf("%x", m.Sum(nil))
	return sign
}
