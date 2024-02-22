package service

import (
	"gorm.io/gorm"
	"server/Service/Ser_LinkUser"
	DB "server/structs/db"
	"time"
)

// 在线列表 数据库处理
type LinksTokenService struct {
	db *gorm.DB
}

// NewLinksTokenService 创建 LinksTokenService 实例
func NewLinksTokenService(db *gorm.DB) *LinksTokenService {
	return &LinksTokenService{
		db: db,
	}
}

// DeleteExpiredTokens 删除已过期的 token
func (s *LinksTokenService) S删除已过期的Token() error {
	// 删除已注销并 6 小时没活动的 token
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Status = 2").Where("LastTime < ?", time.Now().Unix()-21600).Delete("")
	return tx.Error
}

// RevokeExpiredTokens 定时注销已过期的 token
func (s *LinksTokenService) Z注销已过期的Token() error {
	// 注销超时的 token
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Status = 1").Where("LastTime + OutTime < ?", time.Now().Unix()).Updates(map[string]interface{}{"Status": 2, "LogoutCode": Ser_LinkUser.Z注销_心跳超时自动注销})
	return tx.Error
}
