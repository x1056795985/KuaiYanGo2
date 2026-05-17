package service

import (
	"errors"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"time"
)

type S_LogAgentOtherFunc struct{}

func (s *S_LogAgentOtherFunc) Info(tx *gorm.DB, Id int) (db.DB_LogAgentOtherFunc, error) {
	var value db.DB_LogAgentOtherFunc
	err := tx.Model(db.DB_LogAgentOtherFunc{}).Where("Id = ?", Id).First(&value).Error
	return value, err
}

type LogAgentOtherFuncListRequest struct {
	request.List
	Func int64 `json:"Func"` // 操作功能id
}

func (s *S_LogAgentOtherFunc) GetList(tx *gorm.DB, 请求 LogAgentOtherFuncListRequest) (int64, []db.DB_LogAgentOtherFunc, error) {
	局_DB := tx.Model(db.DB_LogAgentOtherFunc{})

	if 请求.Func < 0 {
		局_DB.Where("Func = ?", 请求.Func)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //代理id(用户名转uid)
			// 这里需要查询用户名对应的uid,为避免依赖旧Service,使用直接查询
			var 局_Uid int
			err := tx.Table("db_User").Select("Id").Where("User = ?", 请求.Keywords).Scan(&局_Uid).Error
			if err != nil || 局_Uid == 0 {
				return 0, nil, errors.New("代理账号错误")
			}
			局_DB.Where("AgentUid = ?", 局_Uid)
		case 2: //用户user
			局_DB.Where("AppUser LIKE ?", "%"+请求.Keywords+"%")
		case 3: //ip
			局_DB.Where("Ip LIKE ?", "%"+请求.Keywords+"%")
		case 4: //信息
			局_DB.Where("Note LIKE ?", "%"+请求.Keywords+"%")
		}
	}

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}

	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	var dataList []db.DB_LogAgentOtherFunc
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, dataList, err
}

// BatchDelete 批量删除日志
func (s *S_LogAgentOtherFunc) BatchDelete(tx *gorm.DB, Id []int, Type int, Keywords string) (int64, error) {
	var 影响行数 int64
	var d = tx.Model(db.DB_LogAgentOtherFunc{})

	if Type <= 0 || Type > 7 {
		return 0, errors.New("Type错误")
	}

	switch Type {
	case 1:
		if len(Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		影响行数 = d.Where("Id IN ?", Id).Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 2:
		影响行数 = d.Where("AppUser = ?", Keywords).Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 3:
		影响行数 = d.Where("1=1").Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 4:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-604800).Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 5:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-2592000).Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 6:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-7776000).Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	case 7:
		if len(Keywords) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		影响行数 = d.Where("Note LIKE ?", "%"+Keywords+"%").Delete(db.DB_LogAgentOtherFunc{}).RowsAffected
	}

	if d.Error != nil {
		return 0, d.Error
	}
	return 影响行数, nil
}
