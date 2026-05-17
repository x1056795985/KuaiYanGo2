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

type S_LogLogin struct{}

func (s *S_LogLogin) Info(tx *gorm.DB, Id int) (db.DB_LogLogin, error) {
	var value db.DB_LogLogin
	err := tx.Model(db.DB_LogLogin{}).Where("Id = ?", Id).First(&value).Error
	return value, err
}

func (s *S_LogLogin) GetList(tx *gorm.DB, 请求 request.List, Appid int) (int64, []db.DB_LogLogin, error) {
	局_DB := tx.Model(db.DB_LogLogin{})
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if Appid > 0 {
		局_DB.Where("LoginType = ?", Appid)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("User = ?", 请求.Keywords)
		case 2: //消息
			局_DB.Where("LOCATE(?, Note)>0", 请求.Keywords)
		case 3: //ip
			局_DB.Where("Ip LIKE ?", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	var dataList []db.DB_LogLogin
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, dataList, err
}

// BatchDelete 批量删除日志
// Type: 1删除ID数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前 7删除关键字
func (s *S_LogLogin) BatchDelete(tx *gorm.DB, Id []int, Type int, Keywords string) (int64, error) {
	var 影响行数 int64
	var d = tx.Model(db.DB_LogLogin{})

	if Type <= 0 || Type > 7 {
		return 0, errors.New("Type错误")
	}

	switch Type {
	case 1:
		if len(Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		影响行数 = d.Where("Id IN ?", Id).Delete(db.DB_LogLogin{}).RowsAffected
	case 2:
		影响行数 = d.Where("User = ?", Keywords).Delete(db.DB_LogLogin{}).RowsAffected
	case 3:
		影响行数 = d.Where("1=1").Delete(db.DB_LogLogin{}).RowsAffected
	case 4:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-604800).Delete(db.DB_LogLogin{}).RowsAffected
	case 5:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-2592000).Delete(db.DB_LogLogin{}).RowsAffected
	case 6:
		影响行数 = d.Where("Time < ?", time.Now().Unix()-7776000).Delete(db.DB_LogLogin{}).RowsAffected
	case 7:
		if len(Keywords) == 0 {
			return 0, errors.New("关键字不能为空")
		}
		影响行数 = d.Where("LOCATE(?, Note)>0", Keywords).Delete(db.DB_LogLogin{}).RowsAffected
	}

	if d.Error != nil {
		return 0, d.Error
	}
	return 影响行数, nil
}

// GetAppNameMap 取应用名称映射(用于日志列表附加应用名)
func (s *S_LogLogin) GetAppNameMap(dataList []db.DB_LogLogin) map[string]string {
	// 复用旧架构的AppInfo服务取应用名列表
	var AppNameMap = make(map[string]string)
	// 从全局GVA_DB直接查询,避免依赖旧Service
	type AppName struct {
		Id   int
		Name string
	}
	var apps []AppName
	global.GVA_DB.Table("db_App_Info").Select("Id, Name").Find(&apps)
	for 索引 := range dataList {
		for _, app := range apps {
			if dataList[索引].LoginType == app.Id {
				AppNameMap[strconv.Itoa(dataList[索引].LoginType)] = app.Name
			}
		}
	}
	return AppNameMap
}
