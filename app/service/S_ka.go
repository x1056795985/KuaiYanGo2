package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	dbm "server/app/models/db"
	"strconv"
	"time"
)

const batchSize = 5000

type Ka struct {
	db *gorm.DB
	c  *gin.Context
}

// NewKa 创建 Ka 实例
func NewKa(c *gin.Context, db *gorm.DB) *Ka {
	return &Ka{
		db: db,
		c:  c,
	}
}

func (s *Ka) Info(id int) (info dbm.DB_Ka, err error) {
	tx := s.db.Model(dbm.DB_Ka{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *Ka) InfoKa(Name string) (info dbm.DB_Ka, err error) {
	tx := s.db.Model(dbm.DB_Ka{}).Where("Name = ?", Name).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *Ka) Info2(where map[string]interface{}) (info dbm.DB_Ka, err error) {
	tx := s.db.Model(dbm.DB_Ka{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *Ka) Infos(where map[string]interface{}) (info []dbm.DB_Ka, err error) {
	tx := s.db.Model(dbm.DB_Ka{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *Ka) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_Ka{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *Ka) Delete(Id interface{}) (影响行数 int64, error error) {
	switch k := Id.(type) {
	case int:
		tx := s.db.Model(dbm.DB_Ka{}).Where("Id = ?", k).Delete("")
		return tx.RowsAffected, tx.Error
	case []int:
		var total int64
		for i := 0; i < len(k); i += batchSize {
			end := i + batchSize
			if end > len(k) {
				end = len(k)
			}
			tx := s.db.Model(dbm.DB_Ka{}).Where("Id IN ?", k[i:end]).Delete("")
			if tx.Error != nil {
				return total, tx.Error
			}
			total += tx.RowsAffected
		}
		return total, nil
	default:
		return 0, errors.New("错误的数据")
	}
}

// KaId是否存在
func (s *Ka) KaId是否存在(Appid int, id int) bool {
	var Count int64
	result := s.db.Model(dbm.DB_Ka{}).Select("1").Where("Id=?", id).Where("AppId=?", Appid).First(&Count)
	return result.Error == nil
}

func (s *Ka) Q取总数() int64 {
	var 局_总数 int64
	_ = s.db.Model(dbm.DB_Ka{}).Count(&局_总数).Error
	return 局_总数
}

func (s *Ka) Ka修改状态(id []int, status int) error {
	return s.db.Model(dbm.DB_Ka{}).Where("Id IN ? ", id).Update("Status", status).Error
}

func (s *Ka) Ka修改已用次数加一(id []int) error {
	now := time.Now().Unix()
	return s.db.Model(dbm.DB_Ka{}).
		Where("Id IN ?", id).
		Updates(map[string]interface{}{
			"Num":      gorm.Expr("Num + 1"),
			"UserTime": gorm.Expr("CONCAT(UserTime, ?)", strconv.Itoa(int(now))+","),
			"UseTime":  now,
		}).Error
}

func (s *Ka) Ka修改管理员备注(id []int, AdminNote string) error {
	return s.db.Model(dbm.DB_Ka{}).Where("Id IN ? ", id).Update("AdminNote", AdminNote).Error
}

func (s *Ka) Ka修改代理备注(代理User string, id []int, AgentNote string) error {
	return s.db.Model(dbm.DB_Ka{}).Where("Id IN ? ", id).Where("RegisterUser = ? ", 代理User).Update("AgentNote", AgentNote).Error
}

func (s *Ka) Ka卡号是否存在(卡号 string) bool {
	var Count int64
	_ = s.db.Select("1").Model(dbm.DB_Ka{}).Where("Name=?", 卡号).First(&Count)
	return Count != 0
}

func (s *Ka) Ka卡号取id(Appid int, 卡号 string) int {
	var Id int
	s.db.Model(dbm.DB_Ka{}).Select("Id").Where("Name=?", 卡号).Where("AppId=?", Appid).First(&Id)
	return Id
}

func (s *Ka) Id取制卡人(Id int) string {
	var 制卡人 string
	s.db.Model(dbm.DB_Ka{}).Select("RegisterUser").Where("Id=?", Id).First(&制卡人)
	return 制卡人
}

func (s *Ka) Id检测制卡人(Id []int, 制卡人 string) bool {
	var 实际制卡人 []string
	s.db.Model(dbm.DB_Ka{}).Distinct("RegisterUser").Where("Id IN ?", Id).Find(&实际制卡人)
	if len(实际制卡人) == 1 && 制卡人 == 实际制卡人[0] {
		return true
	}
	return false
}

func (s *Ka) Id取卡号(Id int) string {
	var 卡号 string
	s.db.Model(dbm.DB_Ka{}).Select("Name").Where("Id=?", Id).First(&卡号)
	return 卡号
}

func (s *Ka) Ka卡号取详情(卡号 string) (卡号详情卡号 dbm.DB_Ka, ok error) {
	err := s.db.Model(dbm.DB_Ka{}).Where("Name=?", 卡号).First(&卡号详情卡号).Error
	return 卡号详情卡号, err
}

func (s *Ka) Id取详情(Id int) (卡号详情卡号 dbm.DB_Ka, err error) {
	err = s.db.Model(dbm.DB_Ka{}).Where("Id=?", Id).First(&卡号详情卡号).Error
	return 卡号详情卡号, err
}

func (s *Ka) Ka取已购卡列表(制卡人账号 string, 页数, 数量 int) (卡号详情卡号 []dbm.DB_Ka, ok error) {
	err := s.db.Model(dbm.DB_Ka{}).Order("Id DESC").Where("RegisterUser=?", 制卡人账号).Limit(数量).Offset((页数 - 1) * 数量).Find(&卡号详情卡号).Error
	return 卡号详情卡号, err
}

func (s *Ka) Get卡号总数(AppId int) int {
	var 局_总数 int64
	err := s.db.Model(dbm.DB_Ka{}).Where("AppId=?", AppId).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}

func (s *Ka) Get卡类卡号总数(ClassId int) int {
	var 局_总数 int64
	err := s.db.Model(dbm.DB_Ka{}).Where("KaClassId=?", ClassId).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}

func (s *Ka) Get应用已用和未用数量(AppId int) (已用, 可用 int64) {
	s.db.Model(dbm.DB_Ka{}).Where("AppId=?", AppId).Where("Num=NumMax").Count(&已用)
	s.db.Model(dbm.DB_Ka{}).Where("AppId=?", AppId).Where("Num<NumMax").Count(&可用)
	return
}

func (s *Ka) Get卡类已用和未用数量(卡类Id int) (已用, 可用 int64) {
	s.db.Model(dbm.DB_Ka{}).Where("KaClassId=?", 卡类Id).Where("Num=NumMax").Count(&已用)
	s.db.Model(dbm.DB_Ka{}).Where("KaClassId=?", 卡类Id).Where("Num<NumMax").Count(&可用)
	return
}

// 引用global避免未使用导入
var _ = global.GVA_DB
