package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/logic/admin/L_chart"
	"server/app/logic/admin/L_gaoDe"
	"server/app/models/old/response"
)

type Echart struct {
	Common.Common
}

func NewChartController() *Echart {
	return &Echart{}
}

func (C *Echart) Q取余额消费排行榜(c *gin.Context) {
	var 请求 struct {
		Type int64 `json:"type" binding:"required"`
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
		Type int64 `json:"type" binding:"required"`
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
		Type int64 `json:"type" binding:"required"`
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

// ============ 充值订单图表统计 ============

// Q取充值订单仪表台汇总 充值订单顶部汇总卡片
func (C *Echart) Q取充值订单仪表台汇总(c *gin.Context) {
	data, err := L_chart.Q取充值订单仪表台汇总()
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}

// Q取充值订单分应用月收入统计 分应用本月收入与上月对比
func (C *Echart) Q取充值订单分应用月收入统计(c *gin.Context) {
	data, err := L_chart.Q取充值订单分应用月收入统计()
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}

// Q取充值订单分应用近7天统计 分应用近7天每天成功订单金额
func (C *Echart) Q取充值订单分应用近7天统计(c *gin.Context) {
	data, err := L_chart.Q取充值订单分应用近7天统计()
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}

// Q取充值订单用户充值排行榜 用户充值日/周/月排行榜TOP10
func (C *Echart) Q取充值订单用户充值排行榜(c *gin.Context) {
	var 请求 struct {
		Type int `json:"type"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	data, err := L_chart.Q取充值订单用户充值排行榜(请求.Type)
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}

// Q取充值订单支付方式统计 近30天支付方式金额占比
func (C *Echart) Q取充值订单支付方式统计(c *gin.Context) {
	data, err := L_chart.Q取充值订单支付方式统计()
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}

// Q取充值订单金额区间分布 近30天充值金额区间分布
func (C *Echart) Q取充值订单金额区间分布(c *gin.Context) {
	data, err := L_chart.Q取充值订单金额区间分布()
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "成功", c)
}
