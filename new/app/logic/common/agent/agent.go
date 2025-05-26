package agent

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	m "server/new/app/models/common"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"time"
)

var L_agent agent

func init() {
	L_agent = agent{}

}

type agent struct {
}

// 四舍五入  索引越小,代理级别越靠下
func (j *agent) D代理分成计算(c *gin.Context, 代理id int, 局_总计金额 float64) (局_返回 []m.D代理分成结构, err error) {

	局_返回 = make([]m.D代理分成结构, 0, 3)
	//开始分利润
	var 下级信息 DB.DB_User
	db := *global.GVA_DB
	s_user := service.NewUser(c, &db)
	if 下级信息, err = s_user.Info(代理id); err != nil {
		return 局_返回, fmt.Errorf("代理id:%d,不存在", 代理id)
	}
	局_下级分成百分比 := 0
	for {
		局_临时 := m.D代理分成结构{}
		局_临时.Uid = 下级信息.Id
		局_临时.User = 下级信息.User
		局_临时.F分成百分比 = 下级信息.AgentDiscount
		局_临时.F分给下级百分比 = 局_下级分成百分比
		局_临时.S实际自身百分比 = 下级信息.AgentDiscount - 局_下级分成百分比
		if 局_临时.S实际自身百分比 == 0 {
			局_临时.S实际分成金额 = 0
		} else {
			局_百分比小数 := Float64除int64(D到数值(局_临时.S实际自身百分比), 100, 2) //转换成小数百分比
			局_临时.S实际分成金额 = Float64乘Float64(局_总计金额, 局_百分比小数)
		}

		局_返回 = append(局_返回, 局_临时) //加入到返回数组
		if 下级信息.UPAgentId <= 0 {
			//上级是管理员了 跳出循环
			break
		}

		局_下级分成百分比 = 局_临时.F分成百分比

		if 下级信息, err = s_user.Info(下级信息.UPAgentId); err != nil {
			//代理不存在代理被删了, 结束,返回
			break
		}
		//继续往上找代理
	}
	return 局_返回, nil
}

func (j *agent) Id功能权限检测(c *gin.Context, 代理ID, 权限代号 int) bool {
	var 临时 int
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_卡类授权{}).Select("1").Where("KId=?", 权限代号).Where("Uid=?", 代理ID).Take(&临时)
	return 临时 > 0
}

func (j *agent) S删除代理(c *gin.Context, UID []int) error {
	db := *global.GVA_DB
	err := db.Transaction(func(tx *gorm.DB) error {
		//代理用户删除
		影响行数 := tx.Model(DB.DB_User{}).Where("Id IN ? ", UID).Delete(DB.DB_User{}).RowsAffected
		if 影响行数 == 0 {
			return errors.New("代理用户删除失败")
		}
		// 删除代理关系
		影响行数 = tx.Model(DB.Db_Agent_Level{}).Where("Uid IN ? ", UID).Delete(DB.Db_Agent_Level{}).RowsAffected
		if 影响行数 == 0 {
			return errors.New("代理关系删除失败")
		}
		//删除 代理推广码 NewPromotionCode
		if err := tx.Model(&dbm.DB_PromotionCode{}).Where("Id IN ? ", UID).Delete(&dbm.DB_PromotionCode{}).Error; err != nil {
			return errors.Join(errors.New("代理推广码删除失败"), err)
		}
		//删除代理云配置  appid=50  "AppId": 50,  也不用限制 appid  只要是这个用户的,都删除
		if err := tx.Model(&DB.DB_UserConfig{}).Where("Uid IN ? ", UID).Delete(&DB.DB_UserConfig{}).Error; err != nil {
			return errors.Join(errors.New("代理云配置删除失败"), err)
		}

		return nil
	})

	return err
}

// 不区分用户表还是管理员表
func (j *agent) ID取用户名(c *gin.Context, UID int) (UPAgentName string) {
	db := *global.GVA_DB
	if UID > 0 {
		info, err := service.NewUser(c, &db).Info(UID)
		if err != nil {
			return ""
		}
		return info.User
	} else if UID < 0 {
		info, err := service.NewAdmin(c, &db).Info(-UID)
		if err != nil {
			return ""
		}
		return info.User
	}
	return
}

// 不区分用户表还是管理员表
func (j *agent) ID取分成百分比(c *gin.Context, UID int) (分成百分比 int) {
	db := *global.GVA_DB
	if UID > 0 {
		db.Model(DB.DB_User{}).Select("AgentDiscount").Where("Id=?", UID).First(&分成百分比)
	} else if UID < 0 {
		db.Model(DB.DB_Admin{}).Select("AgentDiscount").Where("Id=?", -UID).First(&分成百分比)
	}
	return
}

// return 可制卡号, 功能权限
func (j *agent) Id取代理可制卡类和可用代理功能列表(c *gin.Context, 代理ID int) ([]int, []int) {

	var 临时 []int
	//不能改下边为  var 可制卡号 []int  否则返回的不是空成员数组,而是nil
	var 可制卡类 = []int{}
	var 功能权限 = []int{}
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_卡类授权{}).Select("Kid").Where("Uid=?", 代理ID).Find(&临时)
	// 将Kid>0的放入可制卡号数组，Kid<0的放入功能权限数组
	for _, kid := range 临时 {
		if kid > 0 {
			可制卡类 = append(可制卡类, kid)
		} else {
			功能权限 = append(功能权限, kid)
		}
	}

	return 可制卡类, 功能权限
}

func (j *agent) D代理授权卡类Id删除(c *gin.Context, kid int) (int64, error) {
	db := *global.GVA_DB
	ret := db.Model(DB.Db_Agent_卡类授权{}).Select("Kid").Where("kid=?", kid).Delete("")

	return ret.RowsAffected, ret.Error
}

func (j *agent) Id取代理可操作应用AppId列表(c *gin.Context, 代理ID int) []int {

	//不能改下边为  var 可制卡号 []int  否则返回的不是空成员数组,而是nil
	var 临时 = []int{}
	var 可制卡类 = []int{}
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_卡类授权{}).Select("Kid").Where("Uid=?", 代理ID).Find(&临时)
	// 将Kid>0的放入可制卡号数组，Kid<0的放入功能权限数组
	for _, kid := range 临时 {
		//排除掉负数 负数是功能,只要卡类ID
		if kid > 0 {
			可制卡类 = append(可制卡类, kid)
		}
	}
	if len(可制卡类) > 0 {
		db.Model(dbm.DB_KaClass{}).Select("AppId").Where("Id IN ?", 可制卡类).Find(&临时)
	}
	临时 = S数组_去重复(临时)
	return 临时
}

// 如果取消了卡类Id,会同时取消下级的该卡类ID
func (j *agent) Z置Id代理可制卡类或功能授权列表(c *gin.Context, 代理ID int, 授权卡类ID []int) error {
	// 查询数据库中代理用户的所有授权卡类ID
	var 已有卡类ID []int
	if err := global.GVA_DB.Model(&DB.Db_Agent_卡类授权{}).Where("Uid = ?", 代理ID).Pluck("KId", &已有卡类ID).Error; err != nil {
		return err
	}
	// 删除数据库中授权卡类ID数组中没有的Kid
	删除卡类ID := S数组_取差集(已有卡类ID, 授权卡类ID)
	if len(删除卡类ID) > 0 {
		if err := global.GVA_DB.Where("Uid = ? AND KId IN ?", 代理ID, 删除卡类ID).Delete(&DB.Db_Agent_卡类授权{}).Error; err != nil {
			return err
		}
	}

	// 增加数据库中,授权卡类ID数组有但数据库没有的Kid
	新增卡类ID := S数组_取差集(授权卡类ID, 已有卡类ID)
	if len(新增卡类ID) > 0 {
		var 新授权记录 []DB.Db_Agent_卡类授权
		for _, 卡类ID := range 新增卡类ID {
			新授权记录 = append(新授权记录, DB.Db_Agent_卡类授权{
				Uid: 代理ID,
				KId: 卡类ID,
			})
		}
		db := *global.GVA_DB
		if err := db.Create(&新授权记录).Error; err != nil {
			return err
		}
	}
	//迭代删除下级中,上级已经被取消的卡类
	j.迭代删除下级代理不允许卡类ID(c, []int{代理ID}, 授权卡类ID)
	return nil
}
func (j *agent) 迭代删除下级代理不允许卡类ID(c *gin.Context, 代理ID []int, 允许卡类ID []int) {
	局_下级代理ID数组 := j.Q取下级代理数组(c, 代理ID)
	if len(局_下级代理ID数组) == 0 { //如果没有下级了,就不继续了
		return
	}
	_ = j.S删除代理不允许使用的卡类(c, 局_下级代理ID数组, 允许卡类ID)
	j.迭代删除下级代理不允许卡类ID(c, 局_下级代理ID数组, 允许卡类ID)
}

func (j *agent) S删除代理不允许使用的卡类(c *gin.Context, 代理ID []int, 允许使用卡类 []int) error {
	db := *global.GVA_DB
	err := db.Where("Uid IN ? ", 代理ID).Where("Kid NOT IN ? ", 允许使用卡类).Delete(&DB.Db_Agent_卡类授权{}).Error
	return err
}

func (j *agent) Q取下级代理数组(c *gin.Context, 上级ID []int) []int {
	var 下级代理 = []int{}
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_Level{}).Select("Uid").Where("UPAgentId IN ?", 上级ID).Where("Level=1").Find(&下级代理)
	return 下级代理
}

func (j *agent) Q取下级代理数组_user(c *gin.Context, 上级ID []int) []string {
	var 局_制卡人数组 = []string{}
	局_数组_uid := j.Q取下级代理数组(c, 上级ID)
	db := *global.GVA_DB
	db.Model(DB.DB_User{}).Select("User").Where("Id IN ?", 局_数组_uid).Find(&局_制卡人数组)
	return 局_制卡人数组

}

func (j *agent) Q取下级代理数组含子级(c *gin.Context, 上级ID []int) []int {
	var 下级代理 = []int{}
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_Level{}).Select("Uid").Where("UPAgentId IN ?", 上级ID).Find(&下级代理)
	return 下级代理
}

// 也可以用来判断是否为上级代理的子级
func (j *agent) Q取上级代理的子级代理级别(c *gin.Context, 上级ID, 子级代理ID int) int {
	var 局_临时整数 = 0
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_Level{}).Select("Level").Where("Uid = ?", 子级代理ID).Where("UPAgentId = ?", 上级ID).Take(&局_临时整数)
	return 局_临时整数
}

// 也可以用来判断是否为上级代理的子级
func (j *agent) S是否都为子级代理(c *gin.Context, 上级ID int, 子级代理ID []int) bool {
	var 局_临时整数 []int
	db := *global.GVA_DB
	db.Model(DB.Db_Agent_Level{}).Select("Level").Where("Uid IN ?", 子级代理ID).Where("UPAgentId = ?", 上级ID).Find(&局_临时整数)
	//查询出来的数量和子级代理ID数量相同,说明,每一个ID,都是子级代理
	return len(局_临时整数) == len(子级代理ID)
}

// 也可以用来判断是否为上级代理的子级
func (j *agent) Q取Id数组中代理数量(c *gin.Context, 代理ID []int) int {
	var 局_临时整数 []int
	db := *global.GVA_DB
	db.Model(DB.DB_User{}).Select("Id").Where("id IN ?", 代理ID).Where("UPAgentId != 0").Find(&局_临时整数)
	//查询出来的数量=0 则说明 没有代理
	return len(局_临时整数)
}

func (j *agent) Id卡类权限检测(c *gin.Context, 代理ID, 卡类ID int) bool {
	return j.Id功能权限检测(c, 代理ID, 卡类ID)
}

func (j *agent) Q取全部代理功能ID_MAP(c *gin.Context) map[int]string {
	局_map := make(map[int]string, 10)
	局_map[DB.D代理功能_卡号冻结] = "卡号冻结"
	局_map[DB.D代理功能_卡号解冻] = "卡号解冻"
	局_map[DB.D代理功能_更换卡号] = "更换卡号"
	//局_map[DB.D代理功能_删除卡号] = "删除卡号"
	局_map[DB.D代理功能_余额充值] = "余额充值"
	局_map[DB.D代理功能_发展下级代理] = "发展下级代理"
	局_map[DB.D代理功能_卡号追回] = "卡号追回"
	局_map[DB.D代理功能_修改用户绑定] = "修改用户绑定"
	局_map[DB.D代理功能_转账] = "转账"
	局_map[DB.D代理功能_代收款] = "代收款"
	局_map[DB.D代理功能_查看归属软件用户] = "查看归属软件用户"
	局_map[DB.D代理功能_冻结软件用户] = "冻结软件用户"
	局_map[DB.D代理功能_解冻软件用户] = "解冻软件用户"
	局_map[DB.D代理功能_修改用户密码] = "修改用户密码"
	局_map[DB.D代理功能_卡类调价] = "卡类调价"
	局_map[DB.D代理功能_制卡] = "制卡"
	return 局_map
}

func (j *agent) Q取全部代理功能名称_MAP(c *gin.Context) map[string]int {
	局_map := j.Q取全部代理功能ID_MAP(c)
	局_map2 := make(map[string]int, len(局_map))
	for key := range 局_map {
		局_map2[局_map[key]] = key
	}
	return 局_map2
}

func (j *agent) Q取全部代理功能ID_int数组(c *gin.Context) []int {
	局_map := j.Q取全部代理功能ID_MAP(c)
	局_数组 := make([]int, 0, len(局_map))
	for key := range 局_map {
		局_数组 = append(局_数组, key)
	}
	return 局_数组
}

func (j *agent) Z执行调价信息分成(c *gin.Context, 调价详情 []dbm.DB_KaClassUpPrice, 购买数量 int64, 日志前缀 string) (err error) {
	if len(调价详情) == 0 {
		return
	}

	var db *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		db = tempObj.(*gorm.DB)
	} else {
		db = &*global.GVA_DB
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		日志记录集 := make([]DB.DB_LogMoney, 0, len(调价详情))

		for _, v := range 调价详情 {
			分成金额 := Float64乘int64(v.Markup, 购买数量) //有多少卡就分多少个
			err = tx.Model(DB.DB_User{}).Where("Id = ?", v.AgentId).Update("RMB", gorm.Expr("RMB + ?", 分成金额)).Error
			if err != nil {
				return errors.Join(err, fmt.Errorf("代理分成失败,请检查原因%d,%s", v.AgentId, Float64到文本(分成金额, 2)))
			}
			var 局_userInfo DB.DB_User
			err = tx.Model(DB.DB_User{}).Where("Id = ?", v.AgentId).Find(&局_userInfo).Error
			if err != nil {
				return errors.Join(err, fmt.Errorf("代理分成后,读取代理数据失败请检查原因%d,%s", v.AgentId, Float64到文本(分成金额, 2)))
			}

			// 构建日志记录
			var 局_临时日志 DB.DB_LogMoney
			局_临时日志.Time = time.Now().Unix()
			局_临时日志.Ip = c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP())
			局_临时日志.User = 局_userInfo.User
			局_临时日志.Count = 分成金额
			局_临时日志.Note = 日志前缀 + fmt.Sprintf("调价分成:¥%s(%s*%d),|新余额≈%s",
				Float64到文本(分成金额, 2),
				Float64到文本(v.Markup, 2),
				购买数量, Float64到文本(局_userInfo.Rmb, 2))
			日志记录集 = append(日志记录集, 局_临时日志)
		}

		// 批量插入日志
		if err = tx.Model(DB.DB_LogMoney{}).Create(&日志记录集).Error; err != nil {
			return fmt.Errorf("日志批量写入失败: %w", err)
		}
		return nil
	})
	return
}
func (j *agent) Z执行百分比代理分成(c *gin.Context, 分成结构 []m.D代理分成结构, 总额度 float64, 日志前缀 string) (err error) {
	if len(分成结构) == 0 {
		return
	}

	var db *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		db = tempObj.(*gorm.DB)
	} else {
		db = &*global.GVA_DB
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		日志记录集 := make([]DB.DB_LogMoney, 0, len(分成结构))

		for _, d := range 分成结构 {
			err = tx.Model(DB.DB_User{}).Where("Id = ?", d.Uid).Update("RMB", gorm.Expr("RMB + ?", d.S实际分成金额)).Error
			if err != nil {
				return errors.Join(err, fmt.Errorf("代理分成失败,请检查原因%d,%s", d.Uid, Float64到文本(d.S实际分成金额, 2)))
			}

			var 局_userInfo DB.DB_User
			err = tx.Model(DB.DB_User{}).Where("Id = ?", d.Uid).Find(&局_userInfo).Error
			if err != nil {
				return errors.Join(err, fmt.Errorf("代理分成后,读取代理数据失败请检查原因%d,%s", d.Uid, Float64到文本(d.S实际分成金额, 2)))
			}
			// 构建日志记录
			var 局_临时日志 DB.DB_LogMoney
			局_临时日志.User = 局_userInfo.User
			局_临时日志.Count = d.S实际分成金额
			局_临时日志.Time = time.Now().Unix()
			局_临时日志.Ip = c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP())
			//分成:¥%s (¥%s*(%d%%-%d%%)),|新余额≈%s",
			局_临时日志.Note = 日志前缀 + fmt.Sprintf(",分成:¥%s (¥%s(实价)*(%d%%-%d%%)),|新余额≈%s",
				Float64到文本(d.S实际分成金额, 2),
				Float64到文本(总额度, 2), d.F分成百分比, d.F分给下级百分比, Float64到文本(局_userInfo.Rmb, 2))
			日志记录集 = append(日志记录集, 局_临时日志)
		}

		// 批量插入日志
		if err = tx.Model(DB.DB_LogMoney{}).Create(&日志记录集).Error; err != nil {
			return fmt.Errorf("日志批量写入失败: %w", err)
		}
		return nil
	})
	return
}
