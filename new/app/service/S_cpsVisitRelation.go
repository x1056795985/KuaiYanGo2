package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsVisitRelation struct {
	*BaseService[dbm.DB_CpsVisitRelation] // 嵌入泛型基础服务
}

// NewCpsVisitRelation 创建 CpsVisitRelation 实例
func NewCpsVisitRelation(c *gin.Context, db *gorm.DB) *CpsVisitRelation {
	return &CpsVisitRelation{
		BaseService: NewBaseService[dbm.DB_CpsVisitRelation](c, db),
	}
}

// 获取邀请人的 所有下级
func (s *CpsVisitRelation) Q取所有被邀请人(指定AppId int, 邀请人id int) (infos []dbm.DB_CpsVisitRelation, err error) {
	return s.Infos(map[string]interface{}{"visitUserId": 邀请人id, "visitedAppId": 指定AppId})
}

func (s *CpsVisitRelation) Q取归属邀请人(指定AppId, 被邀请人id int) (上级, 上上级 dbm.DB_CpsVisitRelation, err error) {
	//预创建 内存变量,防止空指针 通过id 判断是否存在
	上级 = dbm.DB_CpsVisitRelation{}
	上上级 = dbm.DB_CpsVisitRelation{}

	// 查询直接邀请人（一级邀请）
	tx := *s.db
	err = tx.Model(dbm.DB_CpsVisitRelation{}).
		Where("visitedAppId = ?", 指定AppId).
		Where("visitedUserId = ?", 被邀请人id).First(&上级).Error

	if err != nil {
		return
	}

	// 查询间接邀请人（二级邀请）
	err = tx.Model(dbm.DB_CpsVisitRelation{}).
		Where("visitedAppId = ?", 指定AppId).
		Where("visitedUserId = ?", 上级.VisitUserId).First(&上上级).Error

	return
}
