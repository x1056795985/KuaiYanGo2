package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/old/response"
	"strconv"

	"server/app/controller/Common"
	"server/app/service"
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

	if len(请求.Id) == 0 {
		response.FailWithMessage("请选择要删除的调价记录", c)
		return
	}

	tx := *global.GVA_DB
	var S = service.NewKaClassUpPrice(c, &tx)

	//Id去重,防止重复Id导致归属校验误判
	局_Id去重 := make(map[int]struct{}, len(请求.Id))
	局_Id列表 := make([]int, 0, len(请求.Id))
	for _, 局_id := range 请求.Id {
		if _, 局_存在 := 局_Id去重[局_id]; !局_存在 {
			局_Id去重[局_id] = struct{}{}
			局_Id列表 = append(局_Id列表, 局_id)
		}
	}

	//归属校验:只能删除自己的调价记录,防止持权代理删除他人调价记录
	局_自己记录列表, err := S.Infos(map[string]interface{}{"Id": 局_Id列表, "AgentId": c.GetInt("Uid")})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if len(局_自己记录列表) != len(局_Id列表) {
		response.FailWithMessage("包含无权删除的调价记录", c)
		return
	}

	row, err := S.Delete(局_Id列表)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功,数量:"+strconv.Itoa(int(row)), c)

}
