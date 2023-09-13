package DB

//异步任务池,需要三张表

type TaskPool_队列 struct {
	//因为获取队列需要上锁,单独建表提交性能 获取数据不锁数据表
	Uuid string `json:"Uuid" gorm:"column:Uuid;primarykey;size:36;comment:任务结果数据库相同ID"`
	Tid  int    `json:"Tid" gorm:"column:Tid;comment:对应的任务类型id"`
}

func (TaskPool_队列) TableName() string {
	return "db_TaskPoolQueue" //任务队列
}

type TaskPool_数据 struct {
	Uuid string `json:"uuid" gorm:"column:uuid;size:36;primarykey;"`
	//LId        int    `json:"LId" gorm:"column:LId;comment:在线id,只允许相同的查询任务"` 直接用UUid,不可能重复的除了获取者别人也猜不到ID
	Tid        int    `json:"Tid" gorm:"column:Tid;comment:对应的任务类型Id"`
	TimeStart  int    `json:"TimeStart" gorm:"column:TimeStart;comment:任务创建时间戳"`
	TimeEnd    int    `json:"TimeEnd" gorm:"column:TimeEnd;comment:任务结束时间戳"`
	SubmitData string `json:"SubmitData" gorm:"column:SubmitData;comment:生产提交数据"`
	ReturnData string `json:"ReturnData" gorm:"column:ReturnData;comment:消费返回数据"`
	Status     int    `json:"Status" gorm:"column:Status;comment:任务状态,"` //1 已创建,2任务处理中,3成功,4任务失败
}

func (TaskPool_数据) TableName() string {
	return "db_TaskPoolData" //任务数据
}

type TaskPool_数据_精简 struct {
	Uuid       string `json:"uuid" gorm:"column:uuid;size:36;primarykey;"`
	Tid        int    `json:"Tid" gorm:"column:Tid;comment:对应的任务类型Id"`
	TimeStart  int    `json:"TimeStart" gorm:"column:TimeStart;comment:任务创建时间戳"`
	SubmitData string `json:"SubmitData" gorm:"column:SubmitData;comment:生产提交数据"`
}

type TaskPool_类型 struct {
	Id                  int    `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	Name                string `json:"Name" gorm:"column:Name;comment:对应的任务类型名称,也可以当备注"`
	Status              int    `json:"Status" gorm:"column:Status;default:1;comment:任务类型状态 1正常 2维护"`
	HookSubmitDataStart string `json:"HookSubmitDataStart" gorm:"column:HookSubmitDataStart;comment:hook创建入库前函数名"`
	HookSubmitDataEnd   string `json:"HookSubmitDataEnd" gorm:"column:HookSubmitDataEnd;comment:hook创建入库后函数名"`
	HookReturnDataStart string `json:"HookReturnDataStart" gorm:"column:HookReturnDataStart;comment:hook执行入库前函数名"`
	HookReturnDataEnd   string `json:"HookReturnDataEnd" gorm:"column:HookReturnDataEnd;comment:hook执行入库后函数名"`
}

func (TaskPool_类型) TableName() string {
	return "db_TaskPoolType" //任务数据
}
