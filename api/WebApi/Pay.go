package WebApi

import (
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Ser_RMBPayOrder"
	"server/structs/Http/response"
	DB "server/structs/db"
)

func Q取支付订单状态(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"OrderId":[1]}
	局_订单信息 := string(请求json.GetStringBytes("OrderId"))
	if 局_订单信息 == "" { //get也可以
		局_订单信息 = c.Request.FormValue("OrderId")
	}

	if 局_订单信息 == "" {
		response.FailWithMessage("订单不存在", c)
		return
	}
	var 局_订单详细信息 DB.DB_LogRMBPayOrder
	局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(局_订单信息)
	if !ok {
		//如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, ok = Ser_RMBPayOrder.Order取订单详细_第三方订单(局_订单信息)

	}
	if 局_订单详细信息.Id == 0 {
		response.OkWithDetailed([]gin.H{}, "获取成功", c)
		return
	}

	局_响应 := gin.H{"Status": 局_订单详细信息.Status}
	if 局_卡号 := fastjson.GetString([]byte(局_订单详细信息.Extra), "卡号"); 局_卡号 != "" {
		局_响应["KaName"] = 局_卡号
	}
	response.OkWithDetailed(局_响应, "获取成功", c)
	return
}
