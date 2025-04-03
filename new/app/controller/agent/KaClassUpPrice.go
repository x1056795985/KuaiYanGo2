package controller

import (
	"github.com/gin-gonic/gin"
	dbm "server/new/app/models/db"
	"strconv"

	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/service"
	"server/structs/Http/response"
)

type KaClassUpPrice struct {
	Common.Common
}

func NewKaClassUpPriceController() *KaClassUpPrice {
	return &KaClassUpPrice{}
}

func (J *KaClassUpPrice) Save(c *gin.Context) {
	var 请求 struct {
		KaClassId int     `json:"KaClassId" binding:"required,min=1" zh:"卡类"` //校验,必须大于0
		Markup    float64 `json:"Markup" binding:"min=0" zh:"调整价格"`           //校验,必须大于等于0
	}
	//解析失败
	if !J.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewKaClassUpPrice(c, &tx)
	var err error
	var 局 struct {
		调价数据 dbm.DB_KaClassUpPrice
	}

	_, err = service.NewKaClass(c, &tx).Info(请求.KaClassId)
	if err != nil {
		response.FailWithMessage("卡类不存在", c)
		return
	}

	局.调价数据, err = S.Info2(map[string]interface{}{"KaClassId": 请求.KaClassId, "AgentId": c.GetInt("Uid")})
	局.调价数据.AgentId = c.GetInt("Uid")
	局.调价数据.KaClassId = 请求.KaClassId
	局.调价数据.Markup = 请求.Markup
	if 局.调价数据.Id == 0 {
		err = S.Create(&局.调价数据)
	} else {
		_, err = S.Update(局.调价数据.Id, map[string]interface{}{"Markup": 请求.Markup})
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		response.OkWithMessage("操作成功", c)
	}
}

func (J *KaClassUpPrice) Delete(c *gin.Context) {
	var 请求 struct {
		Id []int `json:"Id"`
	}
	//解析失败
	if !J.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewKaClassUpPrice(c, &tx)

	row, err := S.Delete(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithMessage("操作成功,数量:"+strconv.Itoa(int(row)), c)

}
