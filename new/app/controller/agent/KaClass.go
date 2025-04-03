package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/logic/agent/L_KaClass"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/structs/Http/response"
)

type KaClassUp struct {
	Common.Common
}

func NewKaClassController() *KaClassUp {
	return &KaClassUp{}
}

// GetList
func (J *KaClassUp) GetList(c *gin.Context) {
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	var 请求 struct {
		request.List
		AppId int `json:"AppId"`
	}
	if !J.ToJSON(c, &请求) {
		return
	}
	总数, 局_list响应, err := L_KaClass.L_KaClass.GetList(c, 请求.List, 请求.AppId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList{List: 局_list响应, Count: 总数}, "操作成功", c)
	return
}

// 使用预处理数据代替实时计算
func 计算成本价(卡类Id int, 代理层级链 []int, 调价映射 map[int]map[int]float64) float64 {
	// 按层级链顺序查找最近的调价
	for _, agentId := range 代理层级链 {
		if agentMap, exists := 调价映射[agentId]; exists {
			if markup, ok := agentMap[卡类Id]; ok {
				return markup
			}
		}
	}
	// 没有调价则返回卡类基础价格
	return 0 // 从卡类基础信息获取
}
