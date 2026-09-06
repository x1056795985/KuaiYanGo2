package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/models/constant"
	"server/app/models/dbm"
	"server/app/models/request"
	"time"
)

type TaskPoolData struct {
	db *gorm.DB
	c  *gin.Context
}

// NewTaskPoolData 创建 TaskPoolData 实例
func NewTaskPoolData(c *gin.Context, db *gorm.DB) *TaskPoolData {
	return &TaskPoolData{
		db: db,
		c:  c,
	}
}

// 增
func (s *TaskPoolData) Create(info dbm.DB_TaskPoolData) (row int64, err error) {
	tx := s.db.Model(dbm.DB_TaskPoolData{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *TaskPoolData) Delete(Uuid interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Uuid.(type) {
	case string:
		tx2 = s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid = ?", k).Delete("")
	case []string:
		tx2 = s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	//删除缓存
	switch k := Uuid.(type) {
	case string:
		global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + k)
	case []string:
		for _, v := range k {
			global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + v)
		}

	}

	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *TaskPoolData) GetList(请求 request.List, Tid, SubmitAppId, SubmitUid int) (int64, []dbm.DB_TaskPoolData, error) {
	tx := s.db.Model(dbm.DB_TaskPoolData{})
	if Tid > 0 {
		tx = tx.Where("Tid = ?", Tid)
	}

	if SubmitUid > 0 {
		tx = tx.Where("SubmitUid = ?", SubmitUid)
	}

	if SubmitAppId > 0 {
		tx = tx.Where("SubmitAppId = ?", SubmitAppId)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //SubmitData
			tx = tx.Where("SubmitData LIKE ? ", "%"+请求.Keywords+"%")
		case 2: //ReturnData
			tx = tx.Where("ReturnData LIKE ? ", "%"+请求.Keywords+"%")
		case 3: //UUID
			tx = tx.Where("uuid = ? ", 请求.Keywords)
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
		tx = tx.Order("TimeStart ASC")
	case 2:
		tx = tx.Order("TimeStart DESC")
	}
	var 局_数组 []dbm.DB_TaskPoolData
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *TaskPoolData) Info(Uuid string) (info dbm.DB_TaskPoolData, err error) {
	if 局_info, ok := global.H缓存.Get(constant.H缓存前缀_任务池_uuid数据 + Uuid); ok {
		info, ok = 局_info.(dbm.DB_TaskPoolData)
		if ok {
			if info.Status == 3 { //如果任务完成了 则删除缓存一般本次读取后,就不回在轮训了,无需继续缓存数据,即使少量读取,直接读库就行
				global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + Uuid)
			}
			return
		} else {
			global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + Uuid)
		}
	}
	tx := s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).First(&info)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	if info.Status != 3 {
		global.H缓存.Set(constant.H缓存前缀_任务池_uuid数据+Uuid, info, 120*time.Second) //如果有变动只需要缓存120秒
	}
	return
}

// 查
func (s *TaskPoolData) Info2(where map[string]interface{}) (info dbm.DB_TaskPoolData, err error) {
	tx := s.db.Model(dbm.DB_TaskPoolData{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *TaskPoolData) Update(Uuid string, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).Updates(&数据)
	if tx.Error == nil {
		// 先删缓存，再通过 Info2(这个接口不回读取缓存) 回填最新数据，避免旧缓存覆盖新数据
		global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + Uuid)
		info, err2 := s.Info2(map[string]interface{}{"Uuid": Uuid})
		if err2 == nil {
			global.H缓存.Set(constant.H缓存前缀_任务池_uuid数据+Uuid, info, 120*time.Second) //如果有变动只需要缓存120秒
		}
	}

	return tx.RowsAffected, tx.Error
}

// Task数据读取_数组 按Uuid数组取任务数据(精简)
func (s *TaskPoolData) Task数据读取_数组(Uuid []string) []dbm.TaskPool_数据_精简 {
	var TaskPool_数据 []dbm.TaskPool_数据_精简
	if len(Uuid) == 0 {
		return TaskPool_数据
	}
	_ = s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid in ?", Uuid).Find(&TaskPool_数据).Error
	return TaskPool_数据
}

// 按Uuid取Tid
func (s *TaskPoolData) Task数据读取Tid(Uuid string) int {
	if 局_info, ok := global.H缓存.Get(constant.H缓存前缀_任务池_uuid数据 + Uuid); ok {
		局_info2, ok2 := 局_info.(dbm.DB_TaskPoolData)
		if ok2 {
			return 局_info2.Tid
		} else {
			global.H缓存.Delete(constant.H缓存前缀_任务池_uuid数据 + Uuid)
		}
	}
	// 缓存未命中，查库并回填缓存（只查 Tid 不够，需要完整记录才能缓存）
	var info dbm.DB_TaskPoolData
	if err := s.db.Model(dbm.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).First(&info).Error; err == nil {
		global.H缓存.Set(constant.H缓存前缀_任务池_uuid数据+Uuid, info, 120*time.Second)
		return info.Tid
	}
	return 0
}

// Task数据删除过期 删除超过30天的任务
func (s *TaskPoolData) Task数据删除过期() {

	if s.db != nil {
		//删除超过30天的任务
		_ = s.db.Model(dbm.DB_TaskPoolData{}).Where("TimeStart<?", time.Now().Unix()-(86400*30)).Delete("").RowsAffected
		//fmt.Printf("定时删除已过期24H任务:%v\n", 局_数量)
	}
	//24小时
}

var _ = global.GVA_DB // 避免未使用导入
