package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsPayOrder struct {
	*BaseService[dbm.DB_CpsPayOrder] // 嵌入泛型基础服务
}

func NewCpsPayOrder(c *gin.Context, db *gorm.DB) *CpsPayOrder {
	return &CpsPayOrder{
		BaseService: NewBaseService[dbm.DB_CpsPayOrder](c, db),
	}
}
func (s *CpsPayOrder) InfoOrder(order string) (info dbm.DB_CpsPayOrder, err error) {
	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).Where("payOrder = ?", order).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsPayOrder) Q好友订单(appId, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsPayOrder, err error) {

	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).
		Where("appId = ?", appId).
		Where("inviterId = ?", 邀请人id).
		Order("id desc")

	if 数量限制 > 0 {
		tx = tx.Limit(数量限制).
			Find(&infos)
	} else {
		tx = tx.Find(&infos)
	}

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsPayOrder) Q裂变订单(appId, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsPayOrder, err error) {

	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).
		Where("appId = ?", appId).
		Where("grandpaId = ?", 邀请人id).
		Order("id desc")

	if 数量限制 > 0 {
		tx = tx.Limit(数量限制).
			Find(&infos)
	} else {
		tx = tx.Find(&infos)
	}

	if tx.Error != nil {
		err = tx.Error
	}
	return
}
