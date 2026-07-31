package service

import (
	"gorm.io/gorm"
	"server/app/models/constant"
	dbm "server/app/models/db"
)

type S_pay struct {
}

func (s *S_pay) G关闭超时订单(tx *gorm.DB, 过期时间 int64) error {

	db := tx.Model(dbm.DB_LogRMBPayOrder{}).
		Where("Status=?", constant.D订单状态_等待支付).
		Where("Time<?", 过期时间).Update("Status", constant.D订单状态_已关闭)
	return db.Error
}
