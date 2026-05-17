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

type S_LogMoney struct{}

func (s *S_LogMoney) Info(tx *gorm.DB, Id int) (db.DB_LogMoney, error) {
	var value db.DB_LogMoney
	err := tx.Model(db.DB_LogMoney{}).Where("Id = ?", Id).First(&value).Error
	return value, err
}

func (s *S_LogMoney) GetList(tx *gorm.DB, 请求 request.List) (int64, []db.DB_LogMoney, error) {
	局_DB := tx.Model(db.DB_LogMoney{})
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("User = ?", 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE(?, Note)>0", 请求.Keywords)
		case 3: //ip
			局_DB.Where("Ip = ?", 请求.Keywords)
		case 4: //金额
			局_DB.Where("Count = ?", 请求.Keywords)
		}
	}

	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	var dataList []db.DB_LogMoney
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, dataList, err
}

// BatchDelete 批量删除日志
func (s *S_LogMoney) BatchDelete(tx *gorm.DB, Id []int, Type int, Keywords string) (int64, error) {
	var 影响行数 int64
	var d = tx.Model(db.DB_LogMoney{})

	if Type <= 0 || Type > 7 {
		return 0, errors.New("Type错误")
	}

	switch Type {
	case 1:
		if len(Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		影响行数 = d.Where("Id IN ?", Id).Delete(db.DB_LogMoney{}).RowsAffected
	case 2:
		影响行数 = d.Where("User = ?", Keywords).Delete(db.DB_LogMoney{}).RowsAffected
	case 3:
		影响行数 = d.Where("1=1").Delete(db.DB_LogMoney{}).RowsAffected
	case 4:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-604800).Delete(db.DB_LogMoney{}).RowsAffected
	case 5:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-2592000).Delete(db.DB_LogMoney{}).RowsAffected
	case 6:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-7776000).Delete(db.DB_LogMoney{}).RowsAffected
	case 7:
		if len(Keywords) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		影响行数 = d.Where("LOCATE(?, Note)>0", Keywords).Delete(db.DB_LogMoney{}).RowsAffected
	}

	if d.Error != nil {
		return 0, d.Error
	}
	return 影响行数, nil
}

// GetListByUser Agent端按用户过滤的列表查询（带时间范围）
func (s *S_LogMoney) GetListByUser(tx *gorm.DB, 请求 request.ListLog, User string) (int64, []db.DB_LogMoney, error) {
	局_DB := tx.Model(db.DB_LogMoney{}).Where("User = ?", User)
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		制卡结束时间, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_DB.Where("Time > ?", 制卡开始时间).Where("Time < ?", 制卡结束时间+86400)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("User = ?", 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE(?, Note)>0", 请求.Keywords)
		case 3: //ip
			局_DB.Where("Ip = ?", 请求.Keywords)
		}
	}
	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	var dataList []db.DB_LogMoney
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, dataList, err
}

// BatchDeleteByUser Agent端按用户过滤的批量删除
func (s *S_LogMoney) BatchDeleteByUser(tx *gorm.DB, Id []int, Type int, Keywords string, User string) (int64, error) {
	var 影响行数 int64
	var d = tx.Model(db.DB_LogMoney{}).Where("User = ?", User)

	if Type <= 0 || Type > 7 {
		return 0, errors.New("Type错误")
	}

	switch Type {
	case 1:
		if len(Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		影响行数 = d.Where("Id IN ?", Id).Delete(db.DB_LogMoney{}).RowsAffected
	case 2:
		影响行数 = d.Where("User = ?", Keywords).Delete(db.DB_LogMoney{}).RowsAffected
	case 3:
		影响行数 = d.Where("1=1").Delete(db.DB_LogMoney{}).RowsAffected
	case 4:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-604800).Delete(db.DB_LogMoney{}).RowsAffected
	case 5:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-2592000).Delete(db.DB_LogMoney{}).RowsAffected
	case 6:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-7776000).Delete(db.DB_LogMoney{}).RowsAffected
	case 7:
		if len(Keywords) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		影响行数 = d.Where("LOCATE(?, Note)>0", Keywords).Delete(db.DB_LogMoney{}).RowsAffected
	}

	if d.Error != nil {
		return 0, d.Error
	}
	return 影响行数, nil
}
