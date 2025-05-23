package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/logic/admin/L_chart"
	"server/new/app/logic/admin/L_gaoDe"
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

func (C *Echart) G高德取天气(c *gin.Context) {

	data, err := L_gaoDe.G高德查询天气(c)
	if err != nil {
		局_失败提示 := "天气：薛定谔的晴 | 温度：16℃（体感：冰箱冷藏） | 风向：甲方说随便 | 风力：打工人的叹息 | 湿度：60%含泪" //天气：阴 温度：16摄氏度 风向：东 风力：≤3级 空气湿度：60
		response.OkWithDetailed(局_失败提示, "成功", c)
		//response.FailWithMessage("天气读取失败"+err.Error(), c)
	} else {
		response.OkWithDetailed(data, "成功", c)
	}

}
