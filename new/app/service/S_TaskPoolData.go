package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	"server/new/app/models/db"
	"server/new/app/models/request"
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
func (s *TaskPoolData) Create(info db.DB_TaskPoolData) (row int64, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *TaskPoolData) Delete(Uuid interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Uuid.(type) {
	case string:
		tx2 = s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", k).Delete("")
	case []string:
		tx2 = s.db.Model(db.DB_TaskPoolData{}).Where("Uuid IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *TaskPoolData) GetList(请求 request.List, Tid, SubmitAppId, SubmitUid int) (int64, []db.DB_TaskPoolData, error) {
	tx := s.db.Model(db.DB_TaskPoolData{})
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
	var 局_数组 []db.DB_TaskPoolData
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *TaskPoolData) Info(Uuid string) (info db.DB_TaskPoolData, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *TaskPoolData) Info2(where map[string]interface{}) (info db.DB_TaskPoolData, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *TaskPoolData) Update(Uuid string, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 保存
func (s *TaskPoolData) Save(info db.DB_TaskPoolData) (row int64, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", info.Uuid).Save(&info)
	return tx.RowsAffected, tx.Error
}

// Task数据读取_数组 按Uuid数组取任务数据(精简)
func (s *TaskPoolData) Task数据读取_数组(Uuid []string) []db.TaskPool_数据_精简 {
	var TaskPool_数据 []db.TaskPool_数据_精简
	if len(Uuid) == 0 {
		return TaskPool_数据
	}
	_ = s.db.Model(db.DB_TaskPoolData{}).Where("Uuid in ?", Uuid).Find(&TaskPool_数据).Error
	return TaskPool_数据
}

// Task数据读取_单条 按Uuid取单条任务数据
func (s *TaskPoolData) Task数据读取_单条(Uuid string) (db.DB_TaskPoolData, error) {
	var TaskPool_数据 db.DB_TaskPoolData
	err := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).First(&TaskPool_数据).Error
	return TaskPool_数据, err
}

// Task数据读取Tid 按Uuid取Tid
func (s *TaskPoolData) Task数据读取Tid(Uuid string) int {
	var Tid int
	_ = s.db.Model(db.DB_TaskPoolData{}).Select("Tid").Where("Uuid = ?", Uuid).First(&Tid).Error
	return Tid
}

// Task数据修改 数据修改 Status=0 或ReturnData="" 不修改
func (s *TaskPoolData) Task数据修改(Uuid string, Status int, ReturnData string) error {

	局_UpData := make(map[string]interface{}, 3)
	局_UpData["TimeEnd"] = time.Now().Unix()
	if Status != 0 {
		局_UpData["Status"] = Status
	}
	if ReturnData != "" {
		局_UpData["ReturnData"] = ReturnData
	}

	err := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid=?", Uuid).Updates(局_UpData).Error
	return err
}

// Task数据删除过期 删除超过30天的任务
func (s *TaskPoolData) Task数据删除过期() {

	if s.db != nil {
		//删除超过30天的任务
		_ = s.db.Model(db.DB_TaskPoolData{}).Where("TimeStart<?", time.Now().Unix()-(86400*30)).Delete("").RowsAffected
		//fmt.Printf("定时删除已过期24H任务:%v\n", 局_数量)
	}
	//24小时
}

var _ = global.GVA_DB // 避免未使用导入
