package Ser_TaskPool

import (
	"github.com/google/uuid"
	"server/global"
	DB "server/structs/db"
	"strconv"
	"sync"
	"time"
)

var 临界许可 sync.Mutex

func Task队列弹出任务(任务类型id []int, 最大获取数量 int) []string {
	临界许可.Lock()
	defer 临界许可.Unlock()
	var 任务Uuid []string
	if 最大获取数量 == 0 || len(任务类型id) == 0 { //防SB 空信息 还获取  浪费数据库性能
		return 任务Uuid
	}
	db := global.GVA_DB
	_ = db.Model(DB.TaskPool_队列{}).Select("Uuid").Where("Tid in ?", 任务类型id).Limit(最大获取数量).Find(&任务Uuid).Error
	if len(任务类型id) > 0 {
		_ = db.Model(DB.TaskPool_队列{}).Where("Uuid in ?", 任务Uuid).Delete("").Error
		_ = db.Model(DB.TaskPool_数据{}).Where("Uuid in ?", 任务Uuid).Update("Status", 2).Error
	}

	//忽略错误,没有就算了
	return 任务Uuid
}

func Task队列统计Id数量() (map[int]string, error) {

	var results []struct {
		Tid   int
		Count int
	}

	err := global.GVA_DB.Model(DB.TaskPool_队列{}).
		Select("Tid, COUNT(*) AS Count").
		Group("Tid").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	tidCountMap := make(map[int]string)
	for _, result := range results {
		tidCountMap[result.Tid] = strconv.Itoa(result.Count)
	}
	return tidCountMap, nil
}

func Task队列清除指定Tid(Tid []int) (int, error) {
	var 局_uuid []string
	临界许可.Lock()
	defer 临界许可.Unlock()
	global.GVA_DB.Model(DB.TaskPool_队列{}).Select("Uuid").Where("Tid IN ?", Tid).Find(&局_uuid)

	if len(局_uuid) == 0 {
		return 0, nil
	}
	global.GVA_DB.Model(DB.TaskPool_队列{}).Where("Uuid IN ?", 局_uuid).Delete("")

	局_UpData := make(map[string]interface{}, 3)
	局_UpData["TimeEnd"] = time.Now().Unix()
	局_UpData["Status"] = 4

	err := global.GVA_DB.Model(DB.TaskPool_数据{}).Where("Uuid IN ?", 局_uuid).Updates(局_UpData).Error
	if err != nil {
		return 0, err
	}

	return len(局_uuid), nil
}
func Task数据创建加入队列(任务类型Id int, 生产提交数据 string) (string, error) {
	DB_TaskPool_类型 := DB.TaskPool_数据{
		Uuid:       uuid.New().String(),
		Tid:        任务类型Id,
		TimeStart:  int(time.Now().Unix()),
		TimeEnd:    0,
		SubmitData: 生产提交数据,
		ReturnData: "",
		Status:     1,
	}

	err := global.GVA_DB.Model(DB.TaskPool_数据{}).Create(&DB_TaskPool_类型).Error
	if err != nil {
		return "", err
	}

	TaskPool_队列 := DB.TaskPool_队列{
		Uuid: DB_TaskPool_类型.Uuid,
		Tid:  DB_TaskPool_类型.Tid,
	}
	err = global.GVA_DB.Model(DB.TaskPool_队列{}).Create(&TaskPool_队列).Error
	if err != nil {
		//如果失败任务删除丢弃,除非雪崩,不然概率不大,大量出就人工介入
		_ = global.GVA_DB.Model(DB.TaskPool_数据{}).Delete(&DB_TaskPool_类型)
		return "", err
	}

	return DB_TaskPool_类型.Uuid, nil
}
func Task数据读取_数组(Uuid []string) []DB.TaskPool_数据_精简 {
	var TaskPool_数据 []DB.TaskPool_数据_精简
	if len(Uuid) == 0 {
		return TaskPool_数据
	}
	_ = global.GVA_DB.Model(DB.TaskPool_数据{}).Where("Uuid in ?", Uuid).Find(&TaskPool_数据).Error
	return TaskPool_数据
}
func Task数据读取_单条(Uuid string) (DB.TaskPool_数据, error) {
	var TaskPool_数据 DB.TaskPool_数据
	err := global.GVA_DB.Model(DB.TaskPool_数据{}).Where("Uuid = ?", Uuid).First(&TaskPool_数据).Error
	return TaskPool_数据, err
}
func Task数据读取Tid(Uuid string) int {
	var Tid int
	_ = global.GVA_DB.Model(DB.TaskPool_数据{}).Select("Tid").Where("Uuid = ?", Uuid).First(&Tid).Error
	return Tid
}

// 数据修改 Status=0 或ReturnData="" 不修改
func Task数据修改(Uuid string, Status int, ReturnData string) error {

	局_UpData := make(map[string]interface{}, 3)
	局_UpData["TimeEnd"] = time.Now().Unix()
	if Status != 0 {
		局_UpData["Status"] = Status
	}
	if ReturnData != "" {
		局_UpData["ReturnData"] = ReturnData
	}

	err := global.GVA_DB.Model(DB.TaskPool_数据{}).Where("Uuid=?", Uuid).Updates(局_UpData).Error
	return err
}

func Task数据删除过期() {

	if global.GVA_DB != nil {
		//删除超过24小时的任务
		_ = global.GVA_DB.Model(DB.TaskPool_数据{}).Where("TimeStart<?", time.Now().Unix()-86400).Delete("").RowsAffected
		//fmt.Printf("定时删除已过期24H任务:%v\n", 局_数量)
	}
	//24小时
}

func Task类型创建(Name, hook函数名创建入库前, hook函数名创建入库后, hook函数名执行入库前, hook函数名执行入库后, MqttTopicMsg string) error {
	DB_TaskPool_类型 := DB.TaskPool_类型{
		Id:                  0,
		Name:                Name,
		HookSubmitDataStart: hook函数名创建入库前,
		HookSubmitDataEnd:   hook函数名创建入库后,
		HookReturnDataStart: hook函数名执行入库前,
		HookReturnDataEnd:   hook函数名执行入库后,
		MqttTopicMsg:        MqttTopicMsg,
	}
	err := global.GVA_DB.Model(DB.TaskPool_类型{}).Create(&DB_TaskPool_类型).Error
	return err
}

func Task类型修改(id int, Name, hook函数名创建入库前, hook函数名创建入库后, hook函数名创建后, hook函数名执行入库前, hook函数名执行入库后 string) error {
	DB_TaskPool_类型 := DB.TaskPool_类型{
		Id:                  id,
		Name:                Name,
		HookSubmitDataStart: hook函数名创建入库前,
		HookSubmitDataEnd:   hook函数名创建入库后,
		HookReturnDataStart: hook函数名执行入库前,
		HookReturnDataEnd:   hook函数名执行入库后,
	}
	err := global.GVA_DB.Model(DB.TaskPool_类型{}).Save(&DB_TaskPool_类型).Error
	return err
}

func Task类型删除(id int) error {
	err := global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id=?", id).Delete("").Error
	return err
}
func Task类型读取(id int) (DB.TaskPool_类型, error) {
	var DB_TaskPool_类型 DB.TaskPool_类型
	err := global.GVA_DB.Model(DB.TaskPool_类型{}).Where("Id=?", id).First(&DB_TaskPool_类型).Error
	return DB_TaskPool_类型, err
}
