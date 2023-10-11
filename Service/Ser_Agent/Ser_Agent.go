package Ser_Agent

import (
	"errors"
	"gorm.io/gorm"
	"server/Service/Ser_Admin"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
)

// 0 非代理,1 一级代理 2 二级代理 3 三级代理
func Q取Id代理级别(用户ID int) int {
	var Count int64 = 0
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Where("Uid=?", 用户ID).Count(&Count)
	return int(Count)
}

// 第一个成员为三级代理,最后一个成员为 顶级代理
func Q取代理层级信息(userID int) ([]DB.Db_Agent_Level, error) {
	var 数组_代理信息 []DB.Db_Agent_Level
	if Q取Id代理级别(userID) == 0 {
		return 数组_代理信息, nil
	}

	err := 递归获取上级代理ID(userID, &数组_代理信息)
	if err != nil {
		return nil, err
	}
	return 数组_代理信息, nil
}

func 递归获取上级代理ID(userID int, 数组_代理信息 *[]DB.Db_Agent_Level) error {
	var 代理信息 DB.Db_Agent_Level
	err := global.GVA_DB.Where("Uid = ?", userID).First(&代理信息).Error
	if err != nil {
		return err
	}
	*数组_代理信息 = append(*数组_代理信息, 代理信息)
	if 代理信息.UPAgentId < 0 { //如果上级代理小于0 说明已经是管理员了,这个代理为一级代理
		return nil
	}
	return 递归获取上级代理ID(代理信息.UPAgentId, 数组_代理信息)
}

func S删除代理(UID []int) error {
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		影响行数 := tx.Model(DB.DB_User{}).Where("Id IN ? ", UID).Delete("").RowsAffected
		if 影响行数 == 0 {
			return errors.New("代理用户删除失败")
		}
		//代理用户删除了删除代理关系
		影响行数 = tx.Model(DB.Db_Agent_Level{}).Where("Uid IN ? ", UID).Delete("").RowsAffected
		if 影响行数 == 0 {
			return errors.New("代理关系删除失败")
		}
		return nil
	})

	return err
}

// 不区分用户表还是管理员表
func ID取用户名(UID int) (UPAgentName string) {
	if UID > 0 {
		UPAgentName = Ser_User.Id取User(UID)
	} else if UID < 0 {
		UPAgentName = Ser_Admin.Id取User(-UID)
	}
	return
}

// 不区分用户表还是管理员表
func ID取分成百分比(UID int) (分成百分比 int) {
	if UID > 0 {
		_ = global.GVA_DB.Model(DB.DB_User{}).Select("AgentDiscount").Where("Id=?", UID).First(&分成百分比)
	} else if UID < 0 {
		_ = global.GVA_DB.Model(DB.DB_Admin{}).Select("AgentDiscount").Where("Id=?", -UID).First(&分成百分比)
	}
	return
}

// return 可制卡号, 功能权限
func Id取代理可制卡类和可用代理功能列表(代理ID int) ([]int, []int) {

	var 临时 []int
	//不能改下边为  var 可制卡号 []int  否则返回的不是空成员数组,而是nil
	var 可制卡号 = []int{}
	var 功能权限 = []int{}
	global.GVA_DB.Model(DB.Db_Agent_卡类授权{}).Select("Kid").Where("Uid=?", 代理ID).Find(&临时)
	// 将Kid>0的放入可制卡号数组，Kid<0的放入功能权限数组
	for _, kid := range 临时 {
		if kid > 0 {
			可制卡号 = append(可制卡号, kid)
		} else {
			功能权限 = append(功能权限, kid)
		}
	}

	return 可制卡号, 功能权限
}
func Id取代理可操作应用AppId列表(代理ID int) []int {

	//不能改下边为  var 可制卡号 []int  否则返回的不是空成员数组,而是nil
	var 临时 = []int{}
	var 可制卡类 = []int{}
	global.GVA_DB.Model(DB.Db_Agent_卡类授权{}).Select("Kid").Where("Uid=?", 代理ID).Find(&临时)
	// 将Kid>0的放入可制卡号数组，Kid<0的放入功能权限数组
	for _, kid := range 临时 {
		//排除掉负数 负数是功能,只要卡类ID
		if kid > 0 {
			可制卡类 = append(可制卡类, kid)
		}
	}
	if len(可制卡类) > 0 {
		global.GVA_DB.Model(DB.DB_KaClass{}).Select("AppId").Where("Id IN ?", 可制卡类).Find(&临时)
	}
	临时 = 数组_整数去重复(临时)
	return 临时
}
func 数组_整数去重复(arr []int) []int {
	// 创建一个整型key和布尔类型value的哈希表
	hash := make(map[int]bool)
	// 创建一个空的整型数组
	result := []int{}
	// 遍历原始数组
	for _, value := range arr {
		// 如果哈希表(hash)中不存在该值，则加入结果数组和哈希表(hash)
		if _, ok := hash[value]; !ok {
			result = append(result, value)
			hash[value] = true
		}
	}
	// 返回去重后的数组
	return result
}

// 如果取消了卡类Id,会同时取消下级的该卡类ID
func Z置Id代理可制卡类或功能授权列表(代理ID int, 授权卡类ID []int) error {
	// 查询数据库中代理用户的所有授权卡类ID
	var 已有卡类ID []int
	if err := global.GVA_DB.Model(&DB.Db_Agent_卡类授权{}).Where("Uid = ?", 代理ID).Pluck("KId", &已有卡类ID).Error; err != nil {
		return err
	}
	// 删除数据库中授权卡类ID数组中没有的Kid
	删除卡类ID := 差集(已有卡类ID, 授权卡类ID)
	if len(删除卡类ID) > 0 {
		if err := global.GVA_DB.Where("Uid = ? AND KId IN ?", 代理ID, 删除卡类ID).Delete(&DB.Db_Agent_卡类授权{}).Error; err != nil {
			return err
		}
	}

	// 增加数据库中,授权卡类ID数组有但数据库没有的Kid
	新增卡类ID := 差集(授权卡类ID, 已有卡类ID)
	if len(新增卡类ID) > 0 {
		var 新授权记录 []DB.Db_Agent_卡类授权
		for _, 卡类ID := range 新增卡类ID {
			新授权记录 = append(新授权记录, DB.Db_Agent_卡类授权{
				Uid: 代理ID,
				KId: 卡类ID,
			})
		}
		if err := global.GVA_DB.Create(&新授权记录).Error; err != nil {
			return err
		}
	}
	//迭代删除下级中,上级已经被取消的卡类
	迭代删除下级代理不允许卡类ID([]int{代理ID}, 授权卡类ID)
	return nil
}
func 迭代删除下级代理不允许卡类ID(代理ID []int, 允许卡类ID []int) {
	局_下级代理ID数组 := Q取下级代理数组(代理ID)
	if len(局_下级代理ID数组) == 0 { //如果没有下级了,就不继续了
		return
	}
	_ = S删除代理不允许使用的卡类(局_下级代理ID数组, 允许卡类ID)
	迭代删除下级代理不允许卡类ID(局_下级代理ID数组, 允许卡类ID)
}

// 差集函数，返回切片a中有但切片b中没有的元素
func 差集(a, b []int) []int {
	m := make(map[int]bool)
	for _, v := range b {
		m[v] = true
	}

	var 结果 []int
	for _, v := range a {
		if !m[v] {
			结果 = append(结果, v)
		}
	}

	return 结果
}
func S删除代理不允许使用的卡类(代理ID []int, 允许使用卡类 []int) error {
	err := global.GVA_DB.Where("Uid IN ? ", 代理ID).Where("Kid NOT IN ? ", 允许使用卡类).Delete(&DB.Db_Agent_卡类授权{}).Error
	return err
}

func Q取下级代理数组(上级ID []int) []int {
	var 下级代理 = []int{}
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Select("Uid").Where("UPAgentId IN ?", 上级ID).Where("Level=1").Find(&下级代理)
	return 下级代理
}
func Q取下级代理数组含子级(上级ID []int) []int {
	var 下级代理 = []int{}
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Select("Uid").Where("UPAgentId IN ?", 上级ID).Find(&下级代理)
	return 下级代理
}

// 也可以用来判断是否为上级代理的子级
func Q取上级代理的子级代理级别(上级ID, 子级代理ID int) int {
	var 局_临时整数 = 0
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Select("Level").Where("Uid = ?", 子级代理ID).Where("UPAgentId = ?", 上级ID).Take(&局_临时整数)
	return 局_临时整数
}

// 也可以用来判断是否为上级代理的子级
func S是否都为子级代理(上级ID int, 子级代理ID []int) bool {
	var 局_临时整数 []int
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Select("Level").Where("Uid IN ?", 子级代理ID).Where("UPAgentId = ?", 上级ID).Find(&局_临时整数)
	//查询出来的数量和子级代理ID数量相同,说明,每一个ID,都是子级代理
	return len(局_临时整数) == len(子级代理ID)
}

// 也可以用来判断是否为上级代理的子级
func Q取Id数组中代理数量(代理ID []int) int {
	var 局_临时整数 []int
	global.GVA_DB.Model(DB.DB_User{}).Debug().Select("Id").Where("id IN ?", 代理ID).Where("UPAgentId != 0").Find(&局_临时整数)
	//查询出来的数量=0 则说明 没有代理
	return len(局_临时整数)
}
func Id功能权限检测(代理ID, 权限代号 int) bool {
	var 临时 int
	global.GVA_DB.Model(DB.Db_Agent_卡类授权{}).Select("1").Where("KId=?", 权限代号).Where("Uid=?", 代理ID).Take(&临时)
	return 临时 > 0
}

func Id卡类权限检测(代理ID, 卡类ID int) bool {
	return Id功能权限检测(代理ID, 卡类ID)
}

func Q取全部代理功能ID_MAP() map[int]string {
	局_map := make(map[int]string, 10)
	局_map[DB.D代理功能_卡号冻结] = "卡号冻结"
	局_map[DB.D代理功能_卡号解冻] = "卡号解冻"
	局_map[DB.D代理功能_更换卡号] = "更换卡号"
	//局_map[DB.D代理功能_删除卡号] = "删除卡号"
	局_map[DB.D代理功能_余额充值] = "余额充值"
	局_map[DB.D代理功能_发展下级代理] = "发展下级代理"
	局_map[DB.D代理功能_卡号追回] = "卡号追回"
	return 局_map
}
func Q取全部代理功能名称_MAP() map[string]int {
	局_map := Q取全部代理功能ID_MAP()
	局_map2 := make(map[string]int, len(局_map))
	for key := range 局_map {
		局_map2[局_map[key]] = key
	}
	return 局_map2
}

func Q取全部代理功能ID_int数组() []int {
	局_map := Q取全部代理功能ID_MAP()
	局_数组 := make([]int, 0, len(局_map))
	for key := range 局_map {
		局_数组 = append(局_数组, key)
	}
	return 局_数组
}
