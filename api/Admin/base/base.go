// 基础api包  无鉴权  无数据库可以访问
package base

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/structs/Http/response"
	"time"
)

type BaseApi struct{}

// 列宽保存    //已废弃,保存保存在浏览器本地效果更好
func (b *BaseApi) Table宽度保存(c *gin.Context) {
	response.OkWithMessage("已废弃,保存保存在浏览器本地效果更好", c)
	return
	var Request 结构_请求
	err := c.ShouldBindJSON(&Request)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	global.H缓存.Set(Request.Table表名, Request.Table列宽, time.Hour*720) //保存一个月

	response.OkWithMessage("ok", c)

}

// 登录请求结构体
type 结构_请求 struct {
	Table表名 string `json:"Table"`
	Table列宽 []int  `json:"width"` // 列宽数组
}

// 列宽保存
func (b *BaseApi) Table宽度读取(c *gin.Context) {
	response.OkWithMessage("已废弃,保存保存在浏览器本地效果更好", c)
	return
	var Request 结构_请求
	err := c.ShouldBindJSON(&Request)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	局数组_列宽, ok := global.H缓存.Get(Request.Table表名)
	if ok {
		response.OkWithDetailed(局数组_列宽, "ok", c)
	} else {
		response.OkWithDetailed([]int{}, "ok", c)
	}

}
