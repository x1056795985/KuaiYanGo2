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

type S_LogUserMsg struct{}

func (s *S_LogUserMsg) Info(tx *gorm.DB, Id int) (db.DB_LogUserMsg, error) {
	var value db.DB_LogUserMsg
	err := tx.Model(db.DB_LogUserMsg{}).Where("Id = ?", Id).First(&value).Error
	return value, err
}

func (s *S_LogUserMsg) GetList(tx *gorm.DB, 请求 request.ListLog) (int64, []db.DB_LogUserMsg, error) {
	局_DB := tx.Model(db.DB_LogUserMsg{})
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
	if 请求.MsgType > 0 {
		局_DB.Where("MsgType = ?", 请求.MsgType)
	}
	if 请求.AppId > 0 {
		局_DB.Where("AppId = ?", 请求.AppId)
	}
	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	var dataList []db.DB_LogUserMsg
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, dataList, err
}

// BatchDelete 批量删除日志
// Type: 1删除ID数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前 7删除关键字
func (s *S_LogUserMsg) BatchDelete(tx *gorm.DB, Id []int, Type int, Keywords string) (int64, error) {
	var 影响行数 int64
	var d = tx.Model(db.DB_LogUserMsg{})

	if Type <= 0 || Type > 7 {
		return 0, errors.New("Type错误")
	}

	switch Type {
	case 1:
		if len(Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		影响行数 = d.Where("Id IN ?", Id).Delete(db.DB_LogUserMsg{}).RowsAffected
	case 2:
		影响行数 = d.Where("User = ?", Keywords).Delete(db.DB_LogUserMsg{}).RowsAffected
	case 3:
		影响行数 = d.Where("1=1").Delete(db.DB_LogUserMsg{}).RowsAffected
	case 4:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-604800).Delete(db.DB_LogUserMsg{}).RowsAffected
	case 5:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-2592000).Delete(db.DB_LogUserMsg{}).RowsAffected
	case 6:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-7776000).Delete(db.DB_LogUserMsg{}).RowsAffected
	case 7:
		if len(Keywords) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		影响行数 = d.Where("LOCATE(?, Note)>0", Keywords).Delete(db.DB_LogUserMsg{}).RowsAffected
	}

	if d.Error != nil {
		return 0, d.Error
	}
	return 影响行数, nil
}

// SetIsRead 批量修改已读状态
func (s *S_LogUserMsg) SetIsRead(tx *gorm.DB, Id []int, Type int, IsRead bool) error {
	if Type == 1 && len(Id) == 0 {
		return errors.New("Id数组为空")
	}
	if Type == 1 {
		return tx.Model(db.DB_LogUserMsg{}).Where("Id IN ?", Id).Update("IsRead", IsRead).Error
	} else if Type == 2 {
		return tx.Model(db.DB_LogUserMsg{}).Where("1=1").Update("IsRead", IsRead).Error
	}
	return errors.New("操作失败:Type代码错误")
}

// S删除重复消息 删除重复的消息记录
func (s *S_LogUserMsg) S删除重复消息(tx *gorm.DB) error {
	var ids []int
	err := tx.Raw("SELECT min(id) id FROM db_Log_UserMsg GROUP BY Note").Scan(&ids).Error
	if err != nil {
		return err
	}
	err = tx.Debug().Model(db.DB_LogUserMsg{}).Where("id not IN ?", ids).Delete("").Error
	return err
}
