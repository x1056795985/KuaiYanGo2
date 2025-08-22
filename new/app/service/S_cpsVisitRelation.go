package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsInvitingRelation struct {
	*BaseService[dbm.DB_CpsInvitingRelation] // 嵌入泛型基础服务
}

// NewCpsInvitingRelation 创建 CpsInvitingRelation 实例
func NewCpsInvitingRelation(c *gin.Context, db *gorm.DB) *CpsInvitingRelation {
	return &CpsInvitingRelation{
		BaseService: NewBaseService[dbm.DB_CpsInvitingRelation](c, db),
	}
}

// 获取邀请人的 所有下级
func (s *CpsInvitingRelation) Q取所有被邀请人(指定AppId int, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsInvitingRelation, err error) {
	tx := s.db.Model(new(dbm.DB_CpsInvitingRelation)).
		Where(map[string]interface{}{"inviterId": 邀请人id, "inviteeAppId": 指定AppId}).
		Order("id desc").
		Limit(数量限制).
		Find(&infos)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsInvitingRelation) Q取归属邀请人(指定AppId, 被邀请人id int) (上级, 上上级 dbm.DB_CpsInvitingRelation, err error) {
	//预创建 内存变量,防止空指针 通过id 判断是否存在
	上级 = dbm.DB_CpsInvitingRelation{}
	上上级 = dbm.DB_CpsInvitingRelation{}

	// 查询直接邀请人（一级邀请）
	tx := *s.db
	err = tx.Model(dbm.DB_CpsInvitingRelation{}).
		Where("inviteeAppId = ?", 指定AppId).
		Where("inviteeId = ?", 被邀请人id).First(&上级).Error

	if err != nil {
		return
	}

	// 查询间接邀请人（二级邀请）
	err = tx.Model(dbm.DB_CpsInvitingRelation{}).
		Where("inviteeAppId = ?", 指定AppId).
		Where("inviteeId = ?", 上级.InviterId).First(&上上级).Error

	return
}
