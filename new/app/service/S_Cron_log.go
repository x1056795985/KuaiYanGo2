package service

import (
	"errors"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"strconv"
	"time"
)

type S_CronLog struct {
}

// NewCronLogService 创建 NewCronLogService 实例
func NewCronLogService(db *gorm.DB) *S_CronLog {
	return &S_CronLog{}
}

func (s *S_CronLog) Info(tx *gorm.DB, Id int) (db.DB_Cron_log, error) {
	var value db.DB_Cron_log
	err := tx.Model(db.DB_Cron_log{}).Where("Id =?", Id).First(&value).Error
	return value, err
}

func (s *S_CronLog) Update(tx *gorm.DB, value db.DB_Cron_log) error {
	err := tx.Model(db.DB_Cron_log{}).Where("Id = ?", value.Id).Updates(&value).Error
	if err != nil {

	}
	return err
}
func (s *S_CronLog) Create(tx *gorm.DB, value db.DB_Cron_log) error {
	err := tx.Model(db.DB_Cron_log{}).Create(&value).Error
	return err
}

// 删除 支持 数组,和id
func (s *S_CronLog) Delete(tx *gorm.DB, Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = tx.Model(db.DB_Cron_log{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = tx.Model(db.DB_Cron_log{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *S_CronLog) GetList(tx *gorm.DB, 请求 request.List, 结果 int8, Type int, RegisterTime []string) (int64, []db.DB_Cron_log, error) {
	局_DB := tx.Model(db.DB_Cron_log{})

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //任务名称
			局_DB.Where("ReturnText LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	if 结果 > 0 {
		局_DB.Where("Result = ?", 结果)
	}
	if Type > 0 {
		局_DB.Where("Type = ?", Type)
	}
	if RegisterTime != nil && len(RegisterTime) == 2 && RegisterTime[0] != "" && RegisterTime[1] != "" {
		开始时间, _ := strconv.ParseInt(RegisterTime[0], 10, 64)
		结束时间, _ := strconv.ParseInt(RegisterTime[1], 10, 64)
		局_DB.Where("RunTime > ?", 开始时间).Where("RunTime < ?", 结束时间+86400)
	}

	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	//处理排序
	switch 请求.Order {
	default:
		局_DB.Order("Id ASC")
	case 2:
		局_DB.Order("Id DESC")
	}
	var 局_数组 []db.DB_Cron_log
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}

// 批量维护删除
func (s *S_CronLog) DeleteType(tx *gorm.DB, Type int, KeyWord string) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch Type {
	default:
		return 0, errors.New("类型错误")
	case 1: //删除全部
		tx2 = tx.Model(db.DB_Cron_log{}).Where("1=1").Delete("")
	case 2: //删7天前
		tx2 = tx.Model(db.DB_Cron_log{}).Where("RunTime <  ?", time.Now().Unix()-604800).Delete("")
	case 3: //删除30天前
		tx2 = tx.Model(db.DB_Cron_log{}).Where("RunTime <  ?", time.Now().Unix()-2592000).Delete("")
	case 4: //删除90天前
		tx2 = tx.Model(db.DB_Cron_log{}).Where("RunTime <  ?", time.Now().Unix()-7776000).Delete("")
	case 5: //删除关键字
		if len(KeyWord) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		tx2 = tx.Model(db.DB_Cron_log{}).Where("ReturnText like ?", "%"+KeyWord+"%").Delete("")
	}
	return tx2.RowsAffected, tx2.Error
}

// GetAllInfo 获取全部任务信息
func (s *S_CronLog) GetAllInfo(tx *gorm.DB, status int) ([]db.DB_Cron_log, error) {
	var value = []db.DB_Cron_log{}
	var tx2 *gorm.DB
	tx2 = tx.Model(db.DB_Cron_log{})
	if status > 0 {
		tx2 = tx2.Where("Status = ?", status)
	}
	err := tx2.Find(&value).Error

	return value, err
}
