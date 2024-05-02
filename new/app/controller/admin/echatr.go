package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/logic/admin/L_chart"
	"server/structs/Http/response"
)

type Echart struct {
	Common.Common
}

func NewChartController() *Echart {
	return &Echart{}
}

func (C *Echart) Q取余额消费排行榜(c *gin.Context) {
	var 请求 struct {
		Type int64 `json:"Type" binding:"required"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	data, err := L_chart.Q取余额消费排行榜(请求.Type)
	if err != nil {
		return
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(data, "成功", c)
}
func (C *Echart) Q取余额增长排行榜(c *gin.Context) {
	var 请求 struct {
		Type int64 `json:"Type" binding:"required"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	data, err := L_chart.Q取余额增长排行榜(请求.Type)
	if err != nil {
		return
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(data, "成功", c)
}
func (C *Echart) Q取积分消费排行榜(c *gin.Context) {
	var 请求 struct {
		Type int64 `json:"Type" binding:"required"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	data, err := L_chart.Q取积分消费排行榜(请求.Type)
	if err != nil {
		return
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(data, "成功", c)
}
