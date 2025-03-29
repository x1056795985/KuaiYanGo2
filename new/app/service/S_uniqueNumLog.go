package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
	"strconv"
	"time"
)

type UniqueNumLog struct {
	db    *gorm.DB
	c     *gin.Context
	appid int
	table string
}

// NewUniqueNumLog 创建 UniqueNumLog 实例
func NewUniqueNumLog(c *gin.Context, db *gorm.DB, appId int) *UniqueNumLog {
	return &UniqueNumLog{
		db:    db,
		appid: appId,
		c:     c,
		table: dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(appId),
	}
}

// 增
func (s *UniqueNumLog) Create(info *dbm.DB_UniqueNumLog) (row int64, err error) {
	//创建会自动重新赋值info.AppId为新插入的数据AppId
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Create(info)
	return tx.RowsAffected, tx.Error
}

func (s *UniqueNumLog) Info(id int) (info dbm.DB_UniqueNumLog, err error) {
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *UniqueNumLog) InfoItemKey(str string) (info dbm.DB_UniqueNumLog, err error) {

	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where("ItemKey = ?", str).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *UniqueNumLog) Infos(where map[string]interface{}) (info []dbm.DB_UniqueNumLog, err error) {
	info = make([]dbm.DB_UniqueNumLog, 0)
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table + strconv.Itoa(s.appid)).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *UniqueNumLog) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
func (s *UniqueNumLog) Update2(where map[string]interface{}, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where(where).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *UniqueNumLog) Delete(where map[string]interface{}) (影响行数 int64, error error) {
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where(where).Delete("")
	return tx.RowsAffected, tx.Error
}

func (s *UniqueNumLog) Delete已过期() (int64, error) {
	tx := s.db.Model(dbm.DB_UniqueNumLog{}).Table(s.table).Where("EndTime < ?", time.Now().Unix()).Delete("")
	return tx.RowsAffected, tx.Error
}
