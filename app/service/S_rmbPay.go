package service

import (
	"errors"
	"gorm.io/gorm"
	"server/app/global"
	dbm "server/app/models/db"
	"strconv"
	"sync"
	"time"
)

// 在线列表 数据库处理
type RmbPayService struct {
	db *gorm.DB
}

// C处理类型 订单处理类型映射
var C处理类型 = map[int]string{
	0: "余额充值",
	1: "购卡直冲",
	//2: "积分充值",
	3: "支付购卡",
}

// NewRmbPayService 创建 RmbPayService 实例
func NewRmbPayService(db *gorm.DB) *RmbPayService {
	return &RmbPayService{
		db: db,
	}
}

var (
	// 逻辑中使用的某个变量
	集_订单当前秒计数 int
	集_订单当前时间戳 int64
	// 与变量对应的使用互斥锁
	集_互斥锁_订单号 sync.Mutex
)

// Get获取新订单号 生成18位订单号  线程安全
// 年月日时分秒0001计数 每秒9999订单内没问题
func (s *RmbPayService) Get获取新订单号() string {

	集_互斥锁_订单号.Lock()
	当前时间戳 := time.Now().Unix()
	if 当前时间戳 == 集_订单当前时间戳 {
		集_订单当前秒计数++
	} else {
		集_订单当前时间戳 = 当前时间戳
		集_订单当前秒计数 = 1
	}
	局_计数 := 集_订单当前秒计数
	集_互斥锁_订单号.Unlock()

	var 最终订单号 string = time.Unix(当前时间戳, 0).Format("20060102150405")
	if 局_计数 < 10 {
		最终订单号 += "000" + strconv.Itoa(局_计数)
	} else if 局_计数 < 100 {
		最终订单号 += "00" + strconv.Itoa(局_计数)
	} else if 局_计数 < 1000 {
		最终订单号 += "0" + strconv.Itoa(局_计数)
	} else if 局_计数 < 10000 {
		最终订单号 += strconv.Itoa(局_计数)
	} else {
		global.GVA_LOG.Println("恭喜生成订单号大于每秒1w建议更换算法")
	}

	return 最终订单号
}

func (s *RmbPayService) Create(新订单 dbm.DB_LogRMBPayOrder) (dbm.DB_LogRMBPayOrder, error) {

	tx := s.db.Model(dbm.DB_LogRMBPayOrder{}).Create(&新订单)

	if tx.Error != nil {
		return dbm.DB_LogRMBPayOrder{}, tx.Error
	}

	return 新订单, nil
}

func (s *RmbPayService) Info(id int) (info dbm.DB_LogRMBPayOrder, err error) {

	tx := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("Id = ?", id).First(&info)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *RmbPayService) Info2(where map[string]interface{}) (info dbm.DB_LogRMBPayOrder, err error) {

	tx := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where(where).First(&info)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *RmbPayService) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// Order更新订单状态 按订单号更新订单状态
// 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功
func (s *RmbPayService) Order更新订单状态(订单号 string, 状态值 int) bool {
	if 订单号 == "" {
		return false
	}
	err := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Update("Status", 状态值).Error
	if err != nil {
		global.GVA_LOG.Println(订单号 + "Order更新订单状态失败:" + err.Error())
		return false
	}
	return true
}

// Order更新订单备注_批量 批量更新订单备注
func (s *RmbPayService) Order更新订单备注_批量(订单号 []string, 备注 string) error {
	if len(订单号) == 0 {
		return errors.New("订单号数组不能为空")
	}
	err := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("PayOrder IN ?", 订单号).Update("Note", 备注).Error
	return err
}

// Order取订单详细 按订单号取订单详细
func (s *RmbPayService) Order取订单详细(订单号 string) (dbm.DB_LogRMBPayOrder, bool) {
	if 订单号 == "" {
		return dbm.DB_LogRMBPayOrder{}, false
	}
	var 局订单信息 dbm.DB_LogRMBPayOrder
	err := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).First(&局订单信息).Error

	return 局订单信息, err == nil
}

// Order取订单详细_第三方订单 按第三方订单号取订单详细
func (s *RmbPayService) Order取订单详细_第三方订单(第三方订单 string) (dbm.DB_LogRMBPayOrder, bool) {
	if 第三方订单 == "" {
		return dbm.DB_LogRMBPayOrder{}, false
	}
	var 局订单信息 dbm.DB_LogRMBPayOrder
	err := s.db.Model(dbm.DB_LogRMBPayOrder{}).Where("PayOrder2 = ?", 第三方订单).First(&局订单信息).Error

	return 局订单信息, err == nil
}
