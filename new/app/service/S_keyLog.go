package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"strconv"
)

type LogKey struct {
	*BaseService[dbm.DB_LogKey] // 嵌入泛型基础服务
}

// NewLogKey 创建 LogKey 实例
func NewLogKey(c *gin.Context, db *gorm.DB) *LogKey {
	return &LogKey{
		BaseService: NewBaseService[dbm.DB_LogKey](c, db),
	}
}

// 获取列表
func (s *LogKey) GetList(请求 request.List, AppId int, rangeTime []string) (int64, []dbm.DB_LogKey, error) {
	// 创建查询构建器

	局_DB := s.db.Model(dbm.DB_LogKey{})

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名
			局_DB.Where("user = ?", 请求.Keywords)
		case 2: //备注
			局_DB.Where("note LIKE ? ", "%"+请求.Keywords+"%")
		case 3: //绑定信息
			局_DB.Where(s.db.Where("OldKey LIKE ?", "%"+请求.Keywords+"%").
				Or("NewKey LIKE ?", "%"+请求.Keywords+"%"))
		case 4: //ip
			局_DB.Where("ip LIKE ? ", "%"+请求.Keywords+"%")
		}
	}

	if rangeTime != nil && len(rangeTime) == 2 && rangeTime[0] != "" && rangeTime[1] != "" {
		开始时间, _ := strconv.ParseInt(rangeTime[0], 10, 64)
		结束时间, _ := strconv.ParseInt(rangeTime[1], 10, 64)
		局_DB.Where("time > ?", 开始时间).Where("time < ?", 结束时间+86400)
	}

	if AppId >= 10000 {
		局_DB.Where("appId = ?", AppId)
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
	var 局_数组 []dbm.DB_LogKey
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}
