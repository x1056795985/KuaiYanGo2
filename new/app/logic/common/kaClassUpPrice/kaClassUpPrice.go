package kaClassUpPrice

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

var L_kaClassUpPrice = new(kaClassUpPrice)

type kaClassUpPrice struct {
}

// 计算成本价,如果是代理使用,就直接传上级代理id,
func (j *kaClassUpPrice) J计算代理调价(c *gin.Context, 卡类id int, 代理id int) (总加价 float64, 调价信息列表 []dbm.DB_KaClassUpPrice, err error) {

	var info struct {
		代理层级信息 []DB.Db_Agent_Level
		代理Ids  []int
		卡类详情   dbm.DB_KaClass
		临时成本价  float64
	}
	tx := *global.GVA_DB
	info.代理层级信息, err = agentLevel.L_agentLevel.Q取代理层级信息(c, 代理id)
	if err != nil || len(info.代理层级信息) == 0 {
		return
	}
	//需要判断 代理是否有调价权限,因为可能会出现,刚开始有,保存了数据,后续没有了,所以删除了权限,但是数据还在
	// 在循环外部批量获取所有代理的调价信息
	info.代理Ids = make([]int, 0, len(info.代理层级信息))
	for i, _ := range info.代理层级信息 {
		if agent.L_agent.Id功能权限检测(c, info.代理层级信息[i].Uid, DB.D代理功能_卡类调价) {
			info.代理Ids = append(info.代理Ids, info.代理层级信息[i].Uid)
		}
	}
	// 批量查询调价信息（1次查询替代N次）
	调价信息列表, err = service.NewKaClassUpPrice(c, &tx).Infos2(map[string]interface{}{
		"KaClassId": 卡类id,
		"AgentId":   info.代理Ids, // 使用IN查询
	})

	总加价 = 0 // 计算累计调价
	for _, v := range 调价信息列表 {
		总加价 = utils.Float64加float64(总加价, v.Markup, 2)
	}

	return
}
