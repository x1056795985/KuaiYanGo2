package taskPool

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

var L_taskPool taskPool

var 临界许可 sync.Mutex

func init() {
	L_taskPool = taskPool{}
}

type taskPool struct {
}

// Task队列弹出任务 从队列中弹出任务(多表操作: 队列表删除+数据表更新)
func (j *taskPool) Task队列弹出任务(c *gin.Context, 任务类型id []int, 最大获取数量, ReturnAppId, ReturnUid int) []string {
	临界许可.Lock()
	defer 临界许可.Unlock()
	var 任务Uuid []string
	if 最大获取数量 == 0 || len(任务类型id) == 0 { //防SB 空信息 还获取  浪费数据库性能
		return 任务Uuid
	}
	db := global.GVA_DB
	_ = db.Model(dbm.TaskPool_队列{}).Select("Uuid").Where("Tid in ?", 任务类型id).Limit(最大获取数量).Find(&任务Uuid).Error
	if len(任务类型id) > 0 {
		_ = db.Model(dbm.TaskPool_队列{}).Where("Uuid in ?", 任务Uuid).Delete("").Error
		_ = db.Model(dbm.DB_TaskPoolData{}).Where("Uuid in ?", 任务Uuid).Updates(map[string]interface{}{
			"Status":      2,
			"ReturnAppId": ReturnAppId,
			"ReturnUid":   ReturnUid,
		}).Error
	}

	//忽略错误,没有就算了
	return 任务Uuid
}

// Task_取队列数量 取各Tid的队列数量
func (j *taskPool) Task_取队列数量(c *gin.Context) (map[int]string, error) {

	var results []struct {
		Tid   int
		Count int
	}

	err := global.GVA_DB.Model(dbm.TaskPool_队列{}).
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

// Task队列清除指定Tid 清除指定Tid的队列任务(多表操作: 队列表删除+数据表更新)
func (j *taskPool) Task队列清除指定Tid(c *gin.Context, Tid []int) (int, error) {
	var 局_uuid []string
	临界许可.Lock()
	defer 临界许可.Unlock()
	global.GVA_DB.Model(dbm.TaskPool_队列{}).Select("Uuid").Where("Tid IN ?", Tid).Find(&局_uuid)

	if len(局_uuid) == 0 {
		return 0, nil
	}
	global.GVA_DB.Model(dbm.TaskPool_队列{}).Where("Uuid IN ?", 局_uuid).Delete("")

	局_UpData := make(map[string]interface{}, 3)
	局_UpData["TimeEnd"] = time.Now().Unix()
	局_UpData["Status"] = 4

	err := global.GVA_DB.Model(dbm.DB_TaskPoolData{}).Where("Uuid IN ?", 局_uuid).Updates(局_UpData).Error
	if err != nil {
		return 0, err
	}

	return len(局_uuid), nil
}

// Task数据创建加入队列 创建任务数据并加入队列(多表操作: 数据表创建+队列表创建)
func (j *taskPool) Task数据创建加入队列(c *gin.Context, 任务类型Id int, 生产提交数据 string, SubmitAppId, SubmitUid int) (string, error) {
	DB_TaskPool_类型 := dbm.DB_TaskPoolData{
		Uuid:        uuid.New().String(),
		Tid:         任务类型Id,
		TimeStart:   int(time.Now().Unix()),
		TimeEnd:     0,
		SubmitData:  生产提交数据,
		ReturnData:  "",
		Status:      1,
		SubmitAppId: SubmitAppId,
		SubmitUid:   SubmitUid,
	}

	err := global.GVA_DB.Model(dbm.DB_TaskPoolData{}).Create(&DB_TaskPool_类型).Error
	if err != nil {
		return "", err
	}

	TaskPool_队列 := dbm.TaskPool_队列{
		Uuid: DB_TaskPool_类型.Uuid,
		Tid:  DB_TaskPool_类型.Tid,
	}
	err = global.GVA_DB.Model(dbm.TaskPool_队列{}).Create(&TaskPool_队列).Error
	if err != nil {
		//如果失败任务删除丢弃,除非雪崩,不然概率不大,大量出就人工介入
		_ = global.GVA_DB.Model(dbm.DB_TaskPoolData{}).Delete(&DB_TaskPool_类型)
		return "", err
	}

	return DB_TaskPool_类型.Uuid, nil
}

// Uuid_添加到队列 将已有任务Uuid添加到队列(多表操作: 队列表查询+创建)
func (j *taskPool) Uuid_添加到队列(c *gin.Context, uuid string) error {
	var TaskPool_数据 dbm.DB_TaskPoolData
	var TaskPool_队列 dbm.TaskPool_队列
	db := *global.GVA_DB
	//先判断任务是否已经在队列之中

	db.Model(dbm.TaskPool_队列{}).Where("Uuid=?", uuid).First(&TaskPool_队列)
	if TaskPool_队列.Tid != 0 {
		return errors.New("uuid已存在队列之中")
	}
	err := db.Model(dbm.DB_TaskPoolData{}).Where("uuid=?", uuid).First(&TaskPool_数据).Error
	if err != nil {
		return errors.New("uuid任务不存在")
	}
	TaskPool_队列 = dbm.TaskPool_队列{
		Uuid: TaskPool_数据.Uuid,
		Tid:  TaskPool_数据.Tid,
	}
	err = db.Model(dbm.TaskPool_队列{}).Create(&TaskPool_队列).Error
	return err
}

// 确保 gorm 包被使用
var _ = gorm.ErrRecordNotFound
