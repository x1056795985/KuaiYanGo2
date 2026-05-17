package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/publicData"
	"server/structs/Http/response"
)

type PublicDataWebApi struct {
	Common.Common
}

func NewPublicDataWebApiController() *PublicDataWebApi {
	return &PublicDataWebApi{}
}

// Q取公共变量 取公共变量
func (P *PublicDataWebApi) GetPublicData(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文"))
	局_变量名 := string(请求json.GetStringBytes("Name"))
	取值2, err := publicData.L_publicData.Q取值2(c, 1, 局_变量名)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(取值2.Value, "获取成功", c)
	return
}

// Q取队列长度 取公共变量行数
func (P *PublicDataWebApi) GetPublicDataLen(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文"))
	局_变量名 := string(请求json.GetStringBytes("Name"))
	取值2, err := publicData.L_publicData.Q取队列长度(c, 1, 局_变量名)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(取值2, "获取成功", c)
	return
}

// Z置公共变量 置公共变量
func (P *PublicDataWebApi) SetPublicData(c *gin.Context) {
	请求json, _ := fastjson.Parse(c.GetString("局_json明文"))
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
