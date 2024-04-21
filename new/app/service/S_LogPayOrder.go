package service

import (
	"gorm.io/gorm"
	"server/new/app/models/constant"
	DB "server/structs/db"
)

type S_pay struct {
}

func (s *S_pay) G关闭超时订单(tx *gorm.DB, 过期时间 int64) error {

	db := tx.Model(DB.DB_LogRMBPayOrder{}).
		Where("Status=?", constant.D订单状态_等待支付).
		Where("Time<?", 过期时间).Update("Status", constant.D订单状态_已关闭)
	return db.Error
}
