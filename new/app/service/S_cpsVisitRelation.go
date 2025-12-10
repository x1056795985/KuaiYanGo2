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

// 获取邀请人的 所有下级   数量限制=0 无限制
func (s *CpsInvitingRelation) Q取所有被邀请人(指定AppId int, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsInvitingRelation, err error) {
	tx := s.db.Model(new(dbm.DB_CpsInvitingRelation)).
		Where(map[string]interface{}{"inviterId": 邀请人id, "inviteeAppId": 指定AppId}).
		Order("id desc")
	if 数量限制 > 0 {
		tx = tx.Limit(数量限制)
	}
	tx = tx.Find(&infos)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsInvitingRelation) Q取归属邀请人(指定AppId, 被邀请人id int) (上级, 上上级 dbm.DB_CpsInvitingRelation, err error) {
	//预创建 内存变量,防止空指针 通过id 判断是否存在
	var 数组 []dbm.DB_CpsInvitingRelation
	上级 = dbm.DB_CpsInvitingRelation{}
	上上级 = dbm.DB_CpsInvitingRelation{}

	// 查询直接邀请人 含二级

	err = s.db.Model(dbm.DB_CpsInvitingRelation{}).
		Where("inviteeAppId = ?", 指定AppId).
		Where("inviteeId = ?", 被邀请人id).Find(&数组).Error

	if err != nil {
		return
	}

	for _, v := range 数组 {
		if v.Level == 1 {
			上级 = v
		} else if v.Level == 2 {
			上上级 = v
		}
	}

	return
}
