package agentLevel

import (
	"github.com/gin-gonic/gin"
	"server/global"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
)

var L_agentLevel = new(agentLevel)

type agentLevel struct {
}

// 第一个成员为三级代理,最后一个成员为 顶级代理
func (j *agentLevel) Q取代理层级信息(c *gin.Context, userID int) ([]DB.Db_Agent_Level, error) {
	var info struct {
		数组_代理信息 []DB.Db_Agent_Level
		卡类详情    dbm.DB_KaClass
	}
	if j.Q取Id代理级别(c, userID) == 0 {
		return info.数组_代理信息, nil
	}
	//计算耗时

	err := j.递归获取上级代理ID(userID, &info.数组_代理信息)
	if err != nil {
		return nil, err
	}
	return info.数组_代理信息, nil
}

// 0 非代理,1 一级代理 2 二级代理 3 三级代理
func (j *agentLevel) Q取Id代理级别(c *gin.Context, 用户ID int) int {
	var Count int64 = 0
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_Level{}).Where("Uid=?", 用户ID).Count(&Count)
	return int(Count)
}

func (j *agentLevel) 递归获取上级代理ID(userID int, 数组_代理信息 *[]DB.Db_Agent_Level) error {
	var 代理信息 DB.Db_Agent_Level
	db := *global.GVA_DB
	err := db.Where("Uid = ?", userID).First(&代理信息).Error
	if err != nil {
		return err
	}
	*数组_代理信息 = append(*数组_代理信息, 代理信息)
	if 代理信息.UPAgentId < 0 { //如果上级代理小于0 说明已经是管理员了,这个代理为一级代理
		return nil
	}
	return j.递归获取上级代理ID(代理信息.UPAgentId, 数组_代理信息)
}

// 修改递归获取上级代理ID方法为CTE查询  /这种方式虽然只读取一次,但是实际效果,不如索引多次读取速度快,放弃
func (j *agentLevel) 递归获取上级代理ID2(userID int, 数组_代理信息 *[]DB.Db_Agent_Level) error {
	query := `
        WITH RECURSIVE cte AS (
            SELECT Uid, UPAgentId, Level 
            FROM db_Agent_Level 
            WHERE Uid = ?
            UNION ALL
            SELECT a.Uid, a.UPAgentId, a.Level
            FROM db_Agent_Level a
            INNER JOIN cte ON cte.UPAgentId = a.Uid
            WHERE cte.UPAgentId > 0
        )
        SELECT * FROM cte
    `
	db := *global.GVA_DB
	return db.Raw(query, userID).Scan(数组_代理信息).Error
}
