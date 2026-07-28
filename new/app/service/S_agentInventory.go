package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type AgentInventory struct {
	db *gorm.DB
	c  *gin.Context
}

// NewAgentInventory 创建 AgentInventory 实例
func NewAgentInventory(c *gin.Context, db *gorm.DB) *AgentInventory {
	return &AgentInventory{
		db: db,
		c:  c,
	}
}

// Id取详情 按Id取库存卡包详情
func (s *AgentInventory) Id取详情(Id int) (库存卡包详情 dbm.Db_Agent_库存卡包, ok bool) {
	err := s.db.Model(dbm.Db_Agent_库存卡包{}).Where("Id=?", Id).First(&库存卡包详情).Error
	return 库存卡包详情, err == nil
}

// Id取归属Uid 按Id取归属Uid
func (s *AgentInventory) Id取归属Uid(Id int) int {
	if Id == 0 {
		return 0
	}

	Uid := 0
	_ = s.db.Model(dbm.Db_Agent_库存卡包{}).Select("Uid").Where("Id=?", Id).First(&Uid).Error
	return Uid
}

// Id是否存在 库存卡包Id是否存在
func (s *AgentInventory) Id是否存在(Id int) bool {
	var Count int64
	result := s.db.Model(dbm.Db_Agent_库存卡包{}).Select("1").Where("Id=?", Id).First(&Count)
	return result.Error == nil
}

// K库存修改备注 修改库存卡包备注
func (s *AgentInventory) K库存修改备注(库存ID, 代理Uid int, 新备注 string) error {

	原库存详情, ok := s.Id取详情(库存ID)
	if !ok {
		return nil //返回nil保持兼容,实际错误由调用方处理
	}

	if 代理Uid != 原库存详情.Uid {
		return nil
	}

	err := s.db.Model(dbm.Db_Agent_库存卡包{}).Where("Id = ?", 库存ID).Update("Note", 新备注).Error

	return err
}
