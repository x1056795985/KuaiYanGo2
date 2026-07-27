package taskPool

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"server/global"
	DB "server/structs/db"
	"sync"
	"time"
)

var 局_任务队列锁 sync.Mutex

// L_taskPool 任务池 logic (多表操作)
type L_taskPool struct{}

var L_L_taskPool = L_taskPool{}

// Task数据创建加入队列 创建任务数据并加入队列(多表: DB_TaskPoolData + TaskPool_队列)
func (j L_taskPool) Task数据创建加入队列(c *gin.Context, 任务类型Id int, 生产提交数据 string, SubmitAppId, SubmitUid int) (string, error) {
	局_db := *global.GVA_DB
	局_数据 := DB.DB_TaskPoolData{
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

	err := 局_db.Model(DB.DB_TaskPoolData{}).Create(&局_数据).Error
	if err != nil {
		return "", err
	}

	局_队列 := DB.TaskPool_队列{
		Uuid: 局_数据.Uuid,
		Tid:  局_数据.Tid,
	}
	err = 局_db.Model(DB.TaskPool_队列{}).Create(&局_队列).Error
	if err != nil {
		//如果失败任务删除丢弃,除非雪崩,不然概率不大,大量出就人工介入
		_ = 局_db.Model(DB.DB_TaskPoolData{}).Delete(&局_数据)
		return "", err
	}

	return 局_数据.Uuid, nil
}

// Task队列弹出任务 从队列弹出任务(多表: 读队列+删队列+更新TaskPoolData状态)
func (j L_taskPool) Task队列弹出任务(任务类型id []int, 最大获取数量, ReturnAppId, ReturnUid int) []string {
	局_任务队列锁.Lock()
	defer 局_任务队列锁.Unlock()
	var 任务Uuid []string
	if 最大获取数量 == 0 || len(任务类型id) == 0 { //防SB 空信息 还获取  浪费数据库性能
		return 任务Uuid
	}
	db := *global.GVA_DB
	_ = db.Model(DB.TaskPool_队列{}).Select("Uuid").Where("Tid in ?", 任务类型id).Limit(最大获取数量).Find(&任务Uuid).Error
	if len(任务Uuid) > 0 {
		_ = db.Model(DB.TaskPool_队列{}).Where("Uuid in ?", 任务Uuid).Delete("").Error
		_ = db.Model(DB.DB_TaskPoolData{}).Where("Uuid in ?", 任务Uuid).Updates(map[string]interface{}{
			"Status":      2,
			"ReturnAppId": ReturnAppId,
			"ReturnUid":   ReturnUid,
		}).Error
	}

	//忽略错误,没有就算了
	return 任务Uuid
}

// Task数据读取_数组 批量读取任务数据(精简结构)
func (j L_taskPool) Task数据读取_数组(Uuid []string) []DB.TaskPool_数据_精简 {
	var TaskPool_数据 []DB.TaskPool_数据_精简
	if len(Uuid) == 0 {
		return TaskPool_数据
	}
	db := *global.GVA_DB
	_ = db.Model(DB.DB_TaskPoolData{}).Where("Uuid in ?", Uuid).Find(&TaskPool_数据).Error
	return TaskPool_数据
}

// Task数据读取Tid 通过Uuid读取Tid
func (j L_taskPool) Task数据读取Tid(db *gorm.DB, Uuid string) int {
	var Tid int
	_ = db.Model(DB.DB_TaskPoolData{}).Select("Tid").Where("Uuid = ?", Uuid).First(&Tid).Error
	return Tid
}
