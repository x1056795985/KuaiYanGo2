package ka

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/logic/common/agent"
	"server/app/logic/common/kaClassUpPrice"
	dbm "server/app/models/db"
	"server/app/service"
)

// K可制卡类授权树形框结构 可制卡类树形框结构
type K可制卡类授权树形框结构 struct {
	AppId    int    `json:"id"`    //应用AppID
	Label    string `json:"label"` //应用名称
	Children []struct {
		Id    int    `json:"id"`    //卡类Id
		Label string `json:"label"` //卡类名称
	} `json:"children"`
}

// Q取全部可制卡类树形框列表 获取全部可制卡类树形框列表
func (j *ka) Q取全部可制卡类树形框列表(c *gin.Context, 上级代理ID int) []K可制卡类授权树形框结构 {
	局_db := *global.GVA_DB
	var DB_AppInfo []dbm.DB_AppInfo
	_ = global.GVA_DB.Model(dbm.DB_AppInfo{}).Select("AppId", "AppName").Find(&DB_AppInfo).Error

	var DB_KaClass []dbm.DB_KaClass
	if 上级代理ID < 0 {
		//如果小于0说明是开发者,或管理员,可以获取全部卡类   代理价格-1为禁止代理购买的卡号
		_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("AgentMoney>0").Find(&DB_KaClass).Error
	} else {
		var 上级代理可制卡类ID []int
		上级代理可制卡类ID, _ = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 上级代理ID)
		//只可以获取上级代理允许的ID
		_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("Id IN ?", 上级代理可制卡类ID).Where("AgentMoney>0").Find(&DB_KaClass).Error
	}

	var 局_数据 []K可制卡类授权树形框结构
	局_临时上级代理id := service.NewUser(c, &局_db).Id取上级代理ID(上级代理ID)
	for _, app值 := range DB_AppInfo {
		var 局_临时数据 K可制卡类授权树形框结构
		局_临时数据.AppId = 0
		局_临时数据.Label = app值.AppName
		for _, 卡类值 := range DB_KaClass {
			if 卡类值.AppId == app值.AppId {
				if 局_临时上级代理id > 0 {
					局_临时双精度, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 卡类值.Id, 局_临时上级代理id)
					if err == nil && 局_临时双精度 > 0 {
						卡类值.AgentMoney = Float64加float64(卡类值.AgentMoney, 局_临时双精度, 2)
					}
				}

				局_临时数据.Children = append(局_临时数据.Children, struct {
					Id    int    `json:"id"`    //卡类Id
					Label string `json:"label"` //卡类名称
				}{
					Id:    卡类值.Id,
					Label: fmt.Sprintf("Id:%v %v(¥%v)", 卡类值.Id, 卡类值.Name, 卡类值.AgentMoney),
				})

			}
		}
		if len(局_临时数据.Children) > 0 { //只有卡类大于1的应用,才添加进去
			局_数据 = append(局_数据, 局_临时数据)
		}

	}

	return 局_数据
}
