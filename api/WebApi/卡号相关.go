package WebApi

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
)

type 结构请求_单卡号 struct {
	Name string `json:"Name"`
}

// GetKaInfo 获取卡的详细信息

func Get卡号详细信息(c *gin.Context) {
	var 请求 结构请求_单卡号
	//{"Name":"13212315153"}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_Ka DB.DB_Ka

	err = global.GVA_DB.Model(DB.DB_Ka{}).Where("Name = ?", 请求.Name).First(&DB_Ka).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	response.OkWithDetailed(DB_Ka, "获取成功", c)
	return
}
