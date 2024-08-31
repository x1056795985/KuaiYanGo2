package WebApi

import (
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/new/app/logic/common/publicData"
	"server/structs/Http/response"
)

func Q取公共变量(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	取值2, err := publicData.L_publicData.Q取值2(c, 1, 局_变量名)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(取值2.Value, "获取成功", c)
	return
}
func Q取队列长度(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	取值2, err := publicData.L_publicData.Q取队列长度(c, 1, 局_变量名)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(取值2, "获取成功", c)
	return
}
func Z置公共变量(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a","Value":"aaaaa"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_变量值 := string(请求json.GetStringBytes("Value"))

	err := publicData.L_publicData.Z置值(c, 1, 局_变量名, 局_变量值)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
	return
}
