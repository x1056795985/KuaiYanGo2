package rmbWithdraw

import (
	"mime/multipart"

	"gorm.io/gorm"

	dbm "server/app/models/db"
	"server/app/service"
)

// T提现_创建 在事务中创建提现申请并冻结余额。
func T提现_创建(数据库 *gorm.DB, uid int, user string, ip string, 请求 service.WithdrawCreateRequest) (结果 dbm.DB_RmbWithdraw, err error) {
	局_服务 := service.S_RmbWithdraw{}
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		结果, err = 局_服务.Create(tx, uid, user, ip, 请求)
		return err
	})
	return
}

// T提现_取消 在事务中取消提现并退回冻结余额。
func T提现_取消(数据库 *gorm.DB, id int, uid int, user string, ip string) error {
	局_服务 := service.S_RmbWithdraw{}
	return 数据库.Transaction(func(tx *gorm.DB) error {
		return 局_服务.Cancel(tx, id, uid, user, ip)
	})
}

// T提现_审核通过 在事务中将提现单转为付款中。
func T提现_审核通过(数据库 *gorm.DB, id int, adminId int, adminUser string, ip string) error {
	局_服务 := service.S_RmbWithdraw{}
	return 数据库.Transaction(func(tx *gorm.DB) error {
		return 局_服务.AuditPass(tx, id, adminId, adminUser, ip)
	})
}

// T提现_驳回 在事务中驳回提现并退回冻结余额。
func T提现_驳回(数据库 *gorm.DB, id int, reason string, adminId int, adminUser string, ip string) error {
	局_服务 := service.S_RmbWithdraw{}
	return 数据库.Transaction(func(tx *gorm.DB) error {
		return 局_服务.Reject(tx, id, reason, adminId, adminUser, ip)
	})
}

// T提现_标记已付款 在事务中完成提现单。
func T提现_标记已付款(数据库 *gorm.DB, id int, adminId int, adminUser string, ip string) error {
	局_服务 := service.S_RmbWithdraw{}
	return 数据库.Transaction(func(tx *gorm.DB) error {
		return 局_服务.MarkPaid(tx, id, adminId, adminUser, ip)
	})
}

// T提现_上传凭证 在事务中保存凭证记录与操作日志。
func T提现_上传凭证(数据库 *gorm.DB, id int, 文件 *multipart.FileHeader, adminId int, adminUser string, ip string) (路径 string, err error) {
	局_服务 := service.S_RmbWithdraw{}
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		路径, err = 局_服务.UploadVoucher(tx, id, 文件, adminId, adminUser, ip)
		return err
	})
	return
}

// T提现_使用令牌上传凭证 在事务中消费上传令牌并保存凭证。
func T提现_使用令牌上传凭证(数据库 *gorm.DB, token string, 文件 *multipart.FileHeader, ip string) (路径 string, err error) {
	局_服务 := service.S_RmbWithdraw{}
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		路径, err = 局_服务.UploadVoucherByToken(tx, token, 文件, ip)
		return err
	})
	return
}
