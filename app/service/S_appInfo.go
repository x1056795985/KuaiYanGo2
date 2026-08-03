package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/models/request"
	"strconv"
	"time"
)

type AppInfo struct {
	db *gorm.DB
	c  *gin.Context
}

// NewAppInfo 创建 AppInfo 实例
func NewAppInfo(c *gin.Context, db *gorm.DB) *AppInfo {
	return &AppInfo{
		db: db,
		c:  c,
	}
}

// 增
func (s *AppInfo) Create(info dbm.DB_AppInfo) (row int64, err error) {
	//创建会自动重新赋值info.AppId为新插入的数据AppId
	tx := s.db.Model(dbm.DB_AppInfo{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *AppInfo) Delete(AppId interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := AppId.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_AppInfo{}).Where("AppId = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_AppInfo{}).Where("AppId IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *AppInfo) GetList(请求 request.List, Status int) (int64, []dbm.DB_AppInfo, error) {
	tx := s.db
	if Status > 0 {
		tx = tx.Where("Status = ?", Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //AppId
			tx = tx.Where("AppId = ?", 请求.Keywords)
		case 2: //任务名称
			tx = tx.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		tx.Count(&总数)
	}
	//处理排序
	switch 请求.Order {
	default:
		tx = tx.Order("AppId ASC")
	case 2:
		tx = tx.Order("AppId DESC")
	}
	var 局_数组 []dbm.DB_AppInfo
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *AppInfo) Info(AppId int) (info dbm.DB_AppInfo, err error) {
	tx := s.db.Model(dbm.DB_AppInfo{}).Where("AppId = ?", AppId).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *AppInfo) Infos(where map[string]interface{}) (info []dbm.DB_AppInfo, err error) {
	tx := s.db.Model(dbm.DB_AppInfo{}).Where(where).Scan(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *AppInfo) Update(AppId int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_AppInfo{}).Where("AppId = ?", AppId).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// AppInfo取map列表Int 取AppId->AppName的map列表(int键)
func (s *AppInfo) AppInfo取map列表Int(基础id bool) map[int]string {

	var DB_AppInfo []dbm.DB_AppInfo
	var 总数 int64
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppId", "AppName").Count(&总数).Find(&DB_AppInfo).Error
	var AppName = make(map[int]string, 总数+4)
	if 基础id {
		AppName[1] = "管理平台"
		AppName[2] = "代理平台"
		AppName[3] = "WebApi"
		AppName[10] = "WebUser"
		AppName[11] = "WebSocket"
	}

	//吧 id 和 app名字 放入map
	for 索引 := range DB_AppInfo {
		AppName[DB_AppInfo[索引].AppId] = DB_AppInfo[索引].AppName
	}

	return AppName
}

// App取map列表String 取AppId->AppName的map列表(string键)
func (s *AppInfo) App取map列表String(基础id bool) map[string]string {

	局map := s.AppInfo取map列表Int(基础id)
	var 总数 = len(局map)
	var AppName = make(map[string]string, 总数)
	var AppName2 = make(map[string]string, 总数)
	//将map[int]string 转换成 map[string]string
	//需要按键名排序小于10000 的放到最后面
	for key, value := range 局map {
		if key < 10000 {
			AppName2[strconv.Itoa(key)] = value
		} else {
			AppName[strconv.Itoa(key)] = value
		}
	}

	for key, value := range AppName2 {
		AppName[key] = value
	}

	return AppName
}

// App取AppName 按Appid取应用名称
func (s *AppInfo) App取AppName(Appid int) (AppName string) {
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppName").Where("AppId=?", Appid).First(&AppName).Error
	return AppName
}

// App取App详情 按Appid取应用详情(带缓存)
func (s *AppInfo) App取App详情(Appid int) (AppName dbm.DB_AppInfo) {
	Data缓存, ok := global.H缓存.Get("DB_AppInfo_" + strconv.Itoa(Appid)) //读取缓存
	if ok {
		return Data缓存.(dbm.DB_AppInfo)
	}
	_ = s.db.Model(dbm.DB_AppInfo{}).Where("AppId=?", Appid).First(&AppName).Error

	//高频率读取数据 写入缓存
	global.H缓存.Set("DB_AppInfo_"+strconv.Itoa(Appid), AppName, time.Minute*10) //10分钟有效

	return AppName
}

// AppId是否存在 AppId是否存在
func (s *AppInfo) AppId是否存在(AppId int) bool {
	var appInfo int
	result := s.db.Model(dbm.DB_AppInfo{}).Select("1").Where("AppId = ?", AppId).First(&appInfo)
	return result.Error == nil
}

// AppId取应用名称 按AppId取应用名称
func (s *AppInfo) AppId取应用名称(AppId int) string {
	if AppId < 10000 {
		return ""
	}
	AppName := ""
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppName").Where("AppId = ?", AppId).First(&AppName).Error
	return AppName
}

// App取AppType 按Appid取AppType
func (s *AppInfo) App取AppType(Appid int) (AppType int) {
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	return AppType
}

// App是否为卡号 判断应用是否为卡号模式(3=卡号限时,4=卡号计点)
func (s *AppInfo) App是否为卡号(Appid int) bool {
	var AppType int = 0 //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 3 || AppType == 4 {
		return true
	}
	return false
}

// App是否为计点 判断应用是否为计点模式(2=账号计点,4=卡号计点)
func (s *AppInfo) App是否为计点(Appid int) bool {
	var AppType int = 0 //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	_ = s.db.Model(dbm.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 2 || AppType == 4 {
		return true
	}
	return false
}

// App存在数量 按Appid取应用存在数量
func (s *AppInfo) App存在数量(Appid int) int64 {
	var count int64 = 0
	_ = s.db.Model(dbm.DB_AppInfo{}).Where("AppId = ?", Appid).Count(&count).Error

	return count
}
