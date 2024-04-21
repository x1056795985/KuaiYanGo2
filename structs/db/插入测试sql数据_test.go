package DB

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func Test_启动子程序(t *testing.T) {
	/*	for i := 0; i < 99; i++ {
		fmt.Println(Ser_RMBPayOrder.Get获取新订单号())
	}*/
	//插入测试用户()
	//插入用户登录日志()
	//插入余额充值日志()
	//插入余额变化日志()
}
func 连接新数据库() *gorm.DB {

	mysqlConfig := mysql.Config{
		DSN:                       "root:root@tcp(127.0.0.1:3306)/TY?charset=utf8mb4&parseTime=True&loc=Local", // DSN data source name
		DefaultStringSize:         191,                                                                         // string 类型字段的默认长度
		SkipInitializeWithVersion: false,                                                                       // 根据版本自动配置
	}
	旧数据库, err := gorm.Open(mysql.New(mysqlConfig))
	if err != nil {
		//链接失败了
		fmt.Println("就数据库连接失败" + err.Error())
		return nil
	}
	return 旧数据库 //返回连接好的db池
	//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过	global.GVA_DB 获取db对象即可。
	//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接

}
func 连接旧数据库() *gorm.DB {

	mysqlConfig := mysql.Config{
		DSN:                       "root:root@tcp(127.0.0.1:3306)/bsphp?charset=utf8mb4&parseTime=True&loc=Local", // DSN data source name
		DefaultStringSize:         191,                                                                            // string 类型字段的默认长度
		SkipInitializeWithVersion: false,                                                                          // 根据版本自动配置
	}

	旧数据库, err := gorm.Open(mysql.New(mysqlConfig))
	if err != nil {
		//链接失败了
		fmt.Println("新数据库连接失败" + err.Error())
		return nil
	}
	return 旧数据库 //返回连接好的db池
	//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过	global.GVA_DB 获取db对象即可。
	//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接

}
func 插入测试用户() {
	旧数据库 := 连接旧数据库()
	var 局_bsphp_user []bsphp_user
	var Ty_user DB_User
	var 总数 int64
	_ = 旧数据库.Table("bsphp_user").Count(&总数).Find(&局_bsphp_user).Error
	新数据库 := 连接新数据库()
	var err error
	for _, 值 := range 局_bsphp_user {
		Ty_user.Id = 0
		Ty_user.User = 值.UserUser
		Ty_user.PassWord = 值.UserPwd
		Ty_user.Phone = 值.User_Mobile
		Ty_user.Email = 值.UserEmail
		Ty_user.Qq = 值.UserQq
		Ty_user.SuperPassWord = 值.UserMibaoDaan
		Ty_user.Status = int(值.User_IsLock)
		Ty_user.Rmb = 值.UserRmb
		Ty_user.RealNameAttestation = ""
		//Ty_user.Role = int(值.UserDaili)
		Ty_user.UPAgentId, _ = strconv.Atoi(值.UserAngetCarid)
		Ty_user.AgentDiscount = int(值.User_Zhe)

		Ty_user.LoginIp = 值.User_LoginIp
		if Ty_user.LoginIp == "" {
			Ty_user.LoginAppid = 10000 + rand.Intn(5)
		} else {
			Ty_user.LoginAppid = 0
		}
		Ty_user.LoginTime = 值.User_LoginDate.Unix()
		Ty_user.RegisterIp = 值.UserReIp
		Ty_user.RegisterTime = 值.UserReDate.Unix()

		err = 新数据库.Table("db_User").Create(&Ty_user).Error
		if err != nil {
			fmt.Println("插入错误" + err.Error())
		} else {
			//
		}
	}
	fmt.Printf("插用户成功%d", 总数)
	if 旧数据库 != nil || 新数据库 != nil {
		fmt.Println("aaa")
	}
}

type bsphp_user struct {
	UserUid        int
	UserUser       string
	UserPwd        string
	User_IsLock    int8
	UserEmail      string
	UserMailOk     int8
	UserQq         string
	User_Mobile    string
	UserMibaoWenti string
	UserMibaoDaan  string
	UserRmb        float64
	UserJifen      int
	UserReIp       string
	UserReDate     time.Time
	User_LoginIp   string
	User_LoginDate time.Time
	User_CaoShi    int
	UserYao_User   string
	UserYao_Shu    int
	User_Zhe       float64
	User_LoGinNum  int
	User_DenJiTmp  int
	User_DenJi     int
	UserDaili      int8
	UserIsPwd      int    `comment:"是否禁止修改密码"`
	UserBeizhu     string `comment:"用户备注"`
	UserAngetCarid string `comment:"代理卡串"`
}

type BsphpLog struct {
	Id      int
	Leixing string
	Date    int
	Ip      string
	Test    string
	User    string
}

func 插入用户登录日志() {
	旧数据库 := 连接旧数据库()
	var 局_BsphpLog []BsphpLog
	var Ty_user DB_LogLogin
	var 总数 int64
	_ = 旧数据库.Table("bsphp_log").Where("leixing=?", "user_login_log").Count(&总数).Find(&局_BsphpLog).Error
	新数据库 := 连接新数据库()
	角色 := 0
	var err error
	for _, 值 := range 局_BsphpLog {
		Ty_user.Id = 0
		Ty_user.User = 值.User
		Ty_user.Time = int64(值.Date)
		Ty_user.Ip = 值.Ip
		if 值.User == "admin" {
			Ty_user.Note = "管理员登录"
		} else if 值.User == "ssssss" || 值.User == "17835695599" || 值.User == "452068926" {
			Ty_user.Note = "一级代理登录"
		} else {
			Ty_user.Note = "用户登录"
		}

		角色 = 0
		err = 新数据库.Model(局_BsphpLog).Table("db_User").Select("Role").Where("User = ?", 值.User).First(&角色).Error
		if err != nil {
			fmt.Println("查询错误" + err.Error())
		}
		Ty_user.LoginType = 角色

		err = 新数据库.Table("db_log_login").Create(&Ty_user).Error
		if err != nil {
			fmt.Println("插入错误" + err.Error())
		} else {
			//
		}
	}
	fmt.Printf("插用户成功%d", 总数)
	if 旧数据库 != nil || 新数据库 != nil {
		fmt.Println("aaa")
	}
}

/*func 插入余额变化日志() {
	旧数据库 := 连接旧数据库()
	var 局_BsphpLog []BsphpLog
	var Ty_user DB_LogMoney
	var 总数 int64
	_ = 旧数据库.Table("bsphp_log").Where("leixing=?", "money_buy_log").Count(&总数).Find(&局_BsphpLog).Error
	新数据库 := 连接新数据库()
	var err error
	for _, 值 := range 局_BsphpLog {
		Ty_user.Id = 0
		Ty_user.User = 值.User
		Ty_user.Time = 值.Date
		Ty_user.Ip = 值.Ip
		Ty_user.Count, _ = strconv.ParseFloat(W文本_取出中间文本(值.Test+"###", "|", "###"), 64)
		Ty_user.Count = -Ty_user.Count
		Ty_user.Note = 值.Test
		err = 新数据库.Table("db_log_money").Create(&Ty_user).Error

		if err != nil {
			fmt.Println("插入错误" + err.Error())
		} else {
			//
		}
	}
	fmt.Printf("插用户成功" + strconv.FormatInt(总数, 10))
}*/

type BsphpRmbPayLog struct {
	Id           int
	PayId        string
	PayUid       int
	PayLei       string
	PayRbm       float64
	PayZhuangtai int8
	PayInfo1     string
	PayInfo2     string
	PayDate      time.Time
}

func 插入余额充值日志() {
	旧数据库 := 连接旧数据库()
	var 局_BsphpRmbPayLog []BsphpRmbPayLog
	var Ty_user DB_LogRMBPayOrder
	var 总数 int64
	_ = 旧数据库.Table("bsphp_rmb_pay_log").Count(&总数).Find(&局_BsphpRmbPayLog).Error
	新数据库 := 连接新数据库()
	var err error
	for _, 值 := range 局_BsphpRmbPayLog {
		Ty_user.Id = 0
		Ty_user.PayOrder = 值.PayId
		Ty_user.Uid = 值.PayUid
		Ty_user.Status = int(值.PayZhuangtai + 1)
		Ty_user.Type = 值.PayLei
		Ty_user.Rmb = 值.PayRbm
		Ty_user.Time = int64(int(值.PayDate.Unix()))
		Ty_user.Ip = strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255))
		Ty_user.Note = "测试数据自动插入"

		if Ty_user.Type != "zfbdiannao" {
			Ty_user.Note = Ty_user.Type
			Ty_user.Type = "zfbdiannao"
		}

		err = 新数据库.Table("db_log_rmbpayorder").Create(&Ty_user).Error

		if err != nil {
			fmt.Println("插入错误" + err.Error())
		} else {
			//
		}
	}
	fmt.Printf("插用户成功" + strconv.FormatInt(总数, 10))
}
