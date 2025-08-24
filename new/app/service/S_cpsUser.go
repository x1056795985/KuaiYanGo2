package service

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsUser struct {
	*BaseService[dbm.DB_CpsUser] // 嵌入泛型基础服务
}

// NewcpsUser 创建 cpsUser 实例
func NewCpsUser(c *gin.Context, db *gorm.DB) *CpsUser {
	return &CpsUser{
		BaseService: NewBaseService[dbm.DB_CpsUser](c, db),
	}
}

func (s *CpsUser) Z增减累计收入(id int, 增减值 float64, is增加 bool) (err error) {
	err = s.db.Model(new(dbm.DB_CpsUser)).Where("id = ?", id).Update("cumulativeRMB=?", gorm.Expr("cumulativeRMB "+S三元(is增加, "+", "-")+" ?", 增减值)).Error
	return
}
