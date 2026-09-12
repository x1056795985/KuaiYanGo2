package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"server/app/global"
	"server/app/models/constant"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"gorm.io/gorm"

	"server/app/models/dbm"
)

const (
	集_提现配置键名   = "agentWithdrawConfig"
	集_提现凭证令牌前缀 = "withdrawVoucherToken:"
)

type T提现_配置 struct {
	Enable              bool    `json:"enable"`
	MinAmount           float64 `json:"minAmount"`
	MaxAmount           float64 `json:"maxAmount"`
	IntervalSeconds     int64   `json:"intervalSeconds"`
	AllowUserCancel     bool    `json:"allowUserCancel"`
	RequirePayeeQr      bool    `json:"requirePayeeQr"`
	AllowPayeeAccount   bool    `json:"allowPayeeAccount"`
	VoucherTokenSeconds int64   `json:"voucherTokenSeconds"`
	RiskEnable          bool    `json:"riskEnable"`
	PayeeQrMaxSizeMb    int64   `json:"payeeQrMaxSizeMb"`
	VoucherMaxSizeMb    int64   `json:"voucherMaxSizeMb"`
}

type T提现_创建请求 struct {
	Amount         float64 `json:"amount"`
	PayeeType      int     `json:"payeeType"`
	UseLastPayeeQr bool    `json:"useLastPayeeQr"`
	PayeeAccount   string  `json:"payeeAccount"`
	PayeeName      string  `json:"payeeName"`
	UserNote       string  `json:"userNote"`
	RequestId      string  `json:"requestId"`
}

type T提现_列表请求 struct {
	Page         int      `json:"page"`
	Size         int      `json:"size"`
	Status       int      `json:"status"`
	Uid          int      `json:"uid"`
	User         string   `json:"user"`
	OrderNo      string   `json:"orderNo"`
	MinAmount    float64  `json:"minAmount"`
	MaxAmount    float64  `json:"maxAmount"`
	RegisterTime []string `json:"registerTime"`
	Type         int      `json:"type"`
	Keywords     string   `json:"keywords"`
	Count        int64    `json:"count"`
}

type T提现_删除请求 struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

type T提现_图片信息 struct {
	AbsPath string
	Ext     string
}

type T提现_凭证令牌 struct {
	Token        string `json:"token"`
	WithdrawId   int    `json:"withdrawId"`
	AdminId      int    `json:"adminId"`
	AdminUser    string `json:"adminUser"`
	ExpireTime   int64  `json:"expireTime"`
	Used         bool   `json:"used"`
	UploadedPath string `json:"uploadedPath"`
}

type S_RmbWithdraw struct{}

// Q取默认配置 返回提现配置默认值
func Q取默认配置() T提现_配置 {
	return T提现_配置{
		Enable:              false,
		MinAmount:           10,
		MaxAmount:           5000,
		IntervalSeconds:     86400,
		AllowUserCancel:     true,
		RequirePayeeQr:      false,
		AllowPayeeAccount:   true,
		VoucherTokenSeconds: 300,
		RiskEnable:          true,
		PayeeQrMaxSizeMb:    5,
		VoucherMaxSizeMb:    10,
	}
}

func (j *S_RmbWithdraw) Q取配置(数据库 *gorm.DB) T提现_配置 {
	局_配置 := Q取默认配置()
	var 局_设置 dbm.DB_Setting
	if err := 数据库.Model(dbm.DB_Setting{}).Where("ItemKey = ?", 集_提现配置键名).First(&局_设置).Error; err == nil && 局_设置.ItemValue != "" {
		_ = json.Unmarshal([]byte(局_设置.ItemValue), &局_配置)
	}
	if 局_配置.VoucherTokenSeconds <= 0 {
		局_配置.VoucherTokenSeconds = 300
	}
	if 局_配置.PayeeQrMaxSizeMb <= 0 {
		局_配置.PayeeQrMaxSizeMb = 5
	}
	if 局_配置.VoucherMaxSizeMb <= 0 {
		局_配置.VoucherMaxSizeMb = 10
	}
	return 局_配置
}

func (j *S_RmbWithdraw) B保存配置(数据库 *gorm.DB, 配置 T提现_配置) error {
	if 配置.MinAmount < 0 || 配置.MaxAmount < 0 || (配置.MaxAmount > 0 && 配置.MaxAmount < 配置.MinAmount) {
		return errors.New("提现金额配置不正确")
	}
	局_数据, _ := json.Marshal(配置)
	var 局_设置 dbm.DB_Setting
	if err := 数据库.Model(dbm.DB_Setting{}).Where("ItemKey = ?", 集_提现配置键名).First(&局_设置).Error; err == nil {
		return 数据库.Model(dbm.DB_Setting{}).Where("ItemKey = ?", 集_提现配置键名).Update("ItemValue", string(局_数据)).Error
	}
	return 数据库.Model(dbm.DB_Setting{}).Create(&dbm.DB_Setting{ItemKey: 集_提现配置键名, ItemValue: string(局_数据)}).Error
}

func (j *S_RmbWithdraw) Q取代理配置(数据库 *gorm.DB, uid int) (gin.H, error) {
	局_配置 := j.Q取配置(数据库)
	var 局_用户 dbm.DB_User
	if err := 数据库.Model(dbm.DB_User{}).Where("Id = ?", uid).First(&局_用户).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	局_冻结金额 := j.提现_统计金额(数据库, uid, []int{constant.T提现状态_待审核, constant.T提现状态_付款中})
	局_审核中金额 := j.提现_统计金额(数据库, uid, []int{constant.T提现状态_待审核})
	var 局_最近一条 dbm.DB_RmbWithdraw
	_ = 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ?", uid).Order("Id DESC").First(&局_最近一条).Error

	局_下次可提时间 := int64(0)
	if 局_配置.IntervalSeconds > 0 {
		var 局_最近有效 dbm.DB_RmbWithdraw
		if err := 数据库.Model(dbm.DB_RmbWithdraw{}).
			Where("Uid = ? AND Status NOT IN ?", uid, []int{constant.T提现状态_已驳回, constant.T提现状态_已取消}).
			Order("CreateTime DESC").First(&局_最近有效).Error; err == nil {
			局_下次可提时间 = 局_最近有效.CreateTime + 局_配置.IntervalSeconds
			if 局_下次可提时间 < time.Now().Unix() {
				局_下次可提时间 = 0
			}
		}
	}

	局_收款码路径 := 图片_收款码路径(uid)
	_, 局_有收款码 := 文件_存在(路径_转绝对(局_收款码路径))
	return gin.H{
		"enable":             局_配置.Enable,
		"minAmount":          局_配置.MinAmount,
		"maxAmount":          局_配置.MaxAmount,
		"intervalSeconds":    局_配置.IntervalSeconds,
		"allowUserCancel":    局_配置.AllowUserCancel,
		"requirePayeeQr":     局_配置.RequirePayeeQr,
		"allowPayeeAccount":  局_配置.AllowPayeeAccount,
		"nextWithdrawTime":   局_下次可提时间,
		"availableRmb":       局_用户.Rmb,
		"frozenAmount":       局_冻结金额,
		"auditingAmount":     局_审核中金额,
		"lastAmount":         局_最近一条.Amount,
		"lastWithdrawAmount": 局_最近一条.Amount,
		"hasPayeeQr":         局_有收款码,
		"payeeQrPath":        局_收款码路径,
	}, nil
}

func (j *S_RmbWithdraw) S上传收款码(uid int, 文件 *multipart.FileHeader) (string, error) {
	局_数据库 := global.Get局db()
	局_配置 := j.Q取配置(局_数据库)
	return 图片_保存收款码(文件, 图片_收款码路径(uid), 局_配置.PayeeQrMaxSizeMb)
}

func (j *S_RmbWithdraw) Q取代理图片(数据库 *gorm.DB, uid int, 路径 string) (T提现_图片信息, error) {
	局_路径 := 路径_规范图片路径(路径)
	if 局_路径 == "" {
		return T提现_图片信息{}, errors.New("图片地址错误")
	}
	if 局_路径 == 图片_收款码路径(uid) {
		return 图片_取信息(局_路径)
	}

	var 局_数量 int64
	局_错误 := 数据库.Model(dbm.DB_RmbWithdraw{}).
		Where("Uid = ? AND (PayeeQrPath = ? OR VoucherPath = ?)", uid, 局_路径, 局_路径).
		Count(&局_数量).Error
	if 局_错误 != nil {
		return T提现_图片信息{}, 局_错误
	}
	if 局_数量 == 0 {
		return T提现_图片信息{}, errors.New("无权查看该图片")
	}
	return 图片_取信息(局_路径)
}

func (j *S_RmbWithdraw) Q取管理图片(数据库 *gorm.DB, 路径 string) (T提现_图片信息, error) {
	局_路径 := 路径_规范图片路径(路径)
	if 局_路径 == "" {
		return T提现_图片信息{}, errors.New("图片地址错误")
	}

	var 局_数量 int64
	局_错误 := 数据库.Model(dbm.DB_RmbWithdraw{}).
		Where("PayeeQrPath = ? OR VoucherPath = ? OR Id IN (?)", 局_路径, 局_路径,
			数据库.Model(dbm.DB_RmbWithdrawLog{}).Select("WithdrawId").Where("Action IN ? AND LOCATE(?, Note)>0", []int{constant.T提现动作_上传凭证, constant.T提现动作_重新上传凭证}, 凭证_日志路径标记(局_路径))).
		Count(&局_数量).Error
	if 局_错误 != nil {
		return T提现_图片信息{}, 局_错误
	}
	if 局_数量 == 0 {
		return T提现_图片信息{}, errors.New("无权查看该图片")
	}
	return 图片_取信息(局_路径)
}

func (j *S_RmbWithdraw) S上传凭证(数据库 *gorm.DB, 提现Id int, 文件 *multipart.FileHeader, 操作员Id int, 操作员账号 string, ip string) (string, error) {
	局_路径 := fmt.Sprintf("runtime/img/admin/withdraw_voucher_%d_%d%s", 提现Id, time.Now().Unix(), 文件_规范扩展名(文件.Filename))
	局_配置 := j.Q取配置(数据库)
	局_路径, 局_错误 := 图片_保存上传(文件, 局_路径, 局_配置.VoucherMaxSizeMb)
	if 局_错误 != nil {
		return "", 局_错误
	}

	var 局_提现单 dbm.DB_RmbWithdraw
	if 局_错误 = 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", 提现Id).First(&局_提现单).Error; 局_错误 != nil {
		return "", errors.New("提现单不存在")
	}
	if 局_提现单.Status != constant.T提现状态_付款中 && 局_提现单.Status != constant.T提现状态_已付款 {
		return "", errors.New("当前状态不允许上传凭证")
	}

	局_动作 := constant.T提现动作_上传凭证
	if 局_提现单.VoucherPath != "" {
		局_动作 = constant.T提现动作_重新上传凭证
	}
	局_错误 = 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", 提现Id).Updates(map[string]interface{}{
		"VoucherPath":  局_路径,
		"OperatorId":   操作员Id,
		"OperatorUser": 操作员账号,
		"UpdateTime":   time.Now().Unix(),
	}).Error
	if 局_错误 == nil {
		局_错误 = j.提现_写日志(数据库, 局_提现单, 局_提现单.Status, 局_提现单.Status, 局_动作, 操作员Id, 操作员账号, constant.T提现操作人_管理员, ip, "upload voucher "+凭证_日志路径标记(局_路径))
	}
	return 局_路径, 局_错误
}

func (j *S_RmbWithdraw) C创建(数据库 *gorm.DB, uid int, user string, ip string, 请求 T提现_创建请求) (dbm.DB_RmbWithdraw, error) {
	var 局_空单 dbm.DB_RmbWithdraw
	if 请求.Amount <= 0 {
		return 局_空单, errors.New("提现金额必须大于0")
	}

	局_配置 := j.Q取配置(数据库)
	if !局_配置.Enable {
		return 局_空单, errors.New("代理提现未启用")
	}
	if 请求.Amount < 局_配置.MinAmount || (局_配置.MaxAmount > 0 && 请求.Amount > 局_配置.MaxAmount) {
		return 局_空单, errors.New("提现金额不符合规则")
	}
	if 请求.RequestId != "" {
		var 局_旧单 dbm.DB_RmbWithdraw
		if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND RequestId = ?", uid, 请求.RequestId).First(&局_旧单).Error; err == nil {
			return 局_旧单, nil
		}
	}

	var 局_用户 dbm.DB_User
	if err := 数据库.Model(dbm.DB_User{}).Where("Id = ?", uid).First(&局_用户).Error; err != nil {
		return 局_空单, errors.New("用户不存在")
	}
	if 局_用户.Status != 1 {
		return 局_空单, errors.New("用户状态不正常")
	}
	if 局_用户.AgentDiscount <= 0 && 局_用户.UPAgentId == 0 {
		return 局_空单, errors.New("当前账号不是代理")
	}
	if err := j.提现_检查间隔(数据库, uid, 局_配置.IntervalSeconds); err != nil {
		return 局_空单, err
	}

	局_收款码路径 := ""
	if 请求.PayeeType == 1 || 局_配置.RequirePayeeQr {
		局_源路径 := 图片_收款码路径(uid)
		if _, ok := 文件_存在(路径_转绝对(局_源路径)); !ok {
			return 局_空单, errors.New("请先上传收款码")
		}
		局_收款码路径 = fmt.Sprintf("runtime/img/agent/withdraw/payee_qr_%d_%d.jpg", uid, time.Now().Unix())
		if err := 文件_复制(路径_转绝对(局_源路径), 路径_转绝对(局_收款码路径)); err != nil {
			return 局_空单, err
		}
	} else if 请求.PayeeType == 2 {
		if !局_配置.AllowPayeeAccount {
			return 局_空单, errors.New("当前不允许填写收款账号")
		}
		if strings.TrimSpace(请求.PayeeAccount) == "" {
			return 局_空单, errors.New("收款账号不能为空")
		}
	} else {
		return 局_空单, errors.New("收款方式错误")
	}

	局_结果 := 数据库.Model(dbm.DB_User{}).Where("Id = ? AND Rmb >= ?", uid, 请求.Amount).Update("Rmb", gorm.Expr("Rmb - ?", 请求.Amount))
	if 局_结果.Error != nil {
		return 局_空单, 局_结果.Error
	}
	if 局_结果.RowsAffected == 0 {
		return 局_空单, errors.New("余额不足")
	}

	局_时间 := time.Now().Unix()
	局_原始收款信息, _ := json.Marshal(gin.H{"payeeType": 请求.PayeeType, "payeeQrPath": 局_收款码路径, "payeeAccount": 请求.PayeeAccount, "payeeName": 请求.PayeeName})
	局_提现单 := dbm.DB_RmbWithdraw{
		OrderNo:      单号_生成(uid),
		RequestId:    请求.RequestId,
		Uid:          uid,
		User:         user,
		WithdrawType: 1,
		Amount:       请求.Amount,
		Status:       constant.T提现状态_待审核,
		UserNote:     请求.UserNote,
		PayeeType:    请求.PayeeType,
		PayeeQrPath:  局_收款码路径,
		PayeeAccount: 请求.PayeeAccount,
		PayeeName:    请求.PayeeName,
		PayeeRawInfo: string(局_原始收款信息),
		CreateTime:   局_时间,
		UpdateTime:   局_时间,
		Ip:           ip,
	}
	局_错误 := 数据库.Model(dbm.DB_RmbWithdraw{}).Create(&局_提现单).Error
	if 局_错误 == nil {
		局_错误 = j.提现_写资金日志(数据库, uid, user, ip, "提现冻结,提现单号:"+局_提现单.OrderNo, -请求.Amount)
	}
	if 局_错误 == nil {
		局_错误 = j.提现_写日志(数据库, 局_提现单, 0, constant.T提现状态_待审核, constant.T提现动作_提交申请, uid, user, constant.T提现操作人_用户, ip, "提交提现申请")
	}
	return 局_提现单, 局_错误
}

func (j *S_RmbWithdraw) L列表(数据库 *gorm.DB, 请求 T提现_列表请求, uid int) (int64, []dbm.DB_RmbWithdraw, error) {
	请求_规范列表参数(&请求)
	局_查询 := 数据库.Model(dbm.DB_RmbWithdraw{}).Order("Id DESC")
	if uid > 0 {
		局_查询 = 局_查询.Where("Uid = ?", uid)
	}
	if 请求.Status > 0 {
		局_查询 = 局_查询.Where("Status = ?", 请求.Status)
	}
	if 请求.Uid > 0 {
		局_查询 = 局_查询.Where("Uid = ?", 请求.Uid)
	}
	if 请求.User != "" {
		局_查询 = 局_查询.Where("User = ?", 请求.User)
	}
	if 请求.OrderNo != "" {
		局_查询 = 局_查询.Where("OrderNo = ?", 请求.OrderNo)
	}
	if 请求.MinAmount > 0 {
		局_查询 = 局_查询.Where("Amount >= ?", 请求.MinAmount)
	}
	if 请求.MaxAmount > 0 {
		局_查询 = 局_查询.Where("Amount <= ?", 请求.MaxAmount)
	}
	if len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		局_开始, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		局_结束, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_查询 = 局_查询.Where("CreateTime >= ? AND CreateTime < ?", 局_开始, 局_结束+86400)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_查询 = 局_查询.Where("User = ?", 请求.Keywords)
		case 2:
			局_查询 = 局_查询.Where("Uid = ?", 请求.Keywords)
		case 3:
			局_查询 = 局_查询.Where("OrderNo = ?", 请求.Keywords)
		case 4:
			局_查询 = 局_查询.Where("Amount = ?", 请求.Keywords)
		default:
			局_查询 = 局_查询.Where("LOCATE(?, OrderNo)>0 OR LOCATE(?, User)>0", 请求.Keywords, 请求.Keywords)
		}
	}

	var 局_数量 int64
	if 请求.Count > 500000 {
		局_数量 = 请求.Count
	} else {
		局_查询.Count(&局_数量)
	}
	var 局_列表 []dbm.DB_RmbWithdraw
	局_错误 := 局_查询.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_列表).Error
	return 局_数量, 局_列表, 局_错误
}

func (j *S_RmbWithdraw) X详情(数据库 *gorm.DB, id int, uid int) (gin.H, error) {
	var 局_提现单 dbm.DB_RmbWithdraw
	局_查询 := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id)
	if uid > 0 {
		局_查询 = 局_查询.Where("Uid = ?", uid)
	}
	if err := 局_查询.First(&局_提现单).Error; err != nil {
		return nil, errors.New("提现单不存在")
	}

	var 局_日志列表 []dbm.DB_RmbWithdrawLog
	_ = 数据库.Model(dbm.DB_RmbWithdrawLog{}).Where("WithdrawId = ?", id).Order("Id ASC").Find(&局_日志列表).Error
	var 局_用户 dbm.DB_User
	_ = 数据库.Model(dbm.DB_User{}).Where("Id = ?", 局_提现单.Uid).First(&局_用户).Error
	return gin.H{"info": 局_提现单, "logs": 局_日志列表, "user": 局_用户, "riskTags": j.提现_风险标签(数据库, 局_提现单), "voucherHistory": 凭证_历史记录(局_日志列表, 局_提现单.VoucherPath)}, nil
}

func (j *S_RmbWithdraw) Q取消(数据库 *gorm.DB, id int, uid int, user string, ip string) error {
	var 局_提现单 dbm.DB_RmbWithdraw
	if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Uid = ?", id, uid).First(&局_提现单).Error; err != nil {
		return errors.New("提现单不存在")
	}
	if 局_提现单.Status != constant.T提现状态_待审核 {
		return errors.New("只有待审核提现可以取消")
	}
	if !j.Q取配置(数据库).AllowUserCancel {
		return errors.New("当前不允许用户取消提现")
	}
	return j.提现_转退款(数据库, 局_提现单, constant.T提现状态_已取消, constant.T提现动作_用户取消, uid, user, constant.T提现操作人_用户, ip, "用户取消提现")
}

func (j *S_RmbWithdraw) S审核通过(数据库 *gorm.DB, id int, 管理员Id int, 管理员账号 string, ip string) error {
	var 局_提现单 dbm.DB_RmbWithdraw
	if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&局_提现单).Error; err != nil {
		return errors.New("提现单不存在")
	}
	if 局_提现单.Status != constant.T提现状态_待审核 {
		return errors.New("只有待审核提现可以审核通过")
	}
	局_时间 := time.Now().Unix()
	局_错误 := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, constant.T提现状态_待审核).Updates(map[string]interface{}{
		"Status":       constant.T提现状态_付款中,
		"AuditTime":    局_时间,
		"OperatorId":   管理员Id,
		"OperatorUser": 管理员账号,
		"UpdateTime":   局_时间,
	}).Error
	if 局_错误 == nil {
		局_错误 = j.提现_写日志(数据库, 局_提现单, constant.T提现状态_待审核, constant.T提现状态_付款中, constant.T提现动作_审核通过, 管理员Id, 管理员账号, constant.T提现操作人_管理员, ip, "审核通过")
	}
	return 局_错误
}

func (j *S_RmbWithdraw) B驳回(数据库 *gorm.DB, id int, 原因 string, 管理员Id int, 管理员账号 string, ip string) error {
	if strings.TrimSpace(原因) == "" {
		return errors.New("驳回原因不能为空")
	}
	var 局_提现单 dbm.DB_RmbWithdraw
	if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&局_提现单).Error; err != nil {
		return errors.New("提现单不存在")
	}
	if 局_提现单.Status != constant.T提现状态_待审核 && 局_提现单.Status != constant.T提现状态_付款中 {
		return errors.New("当前状态不允许驳回")
	}
	局_动作 := constant.T提现动作_驳回
	if 局_提现单.Status == constant.T提现状态_付款中 {
		局_动作 = constant.T提现动作_付款失败驳回
	}
	return j.提现_转退款(数据库, 局_提现单, constant.T提现状态_已驳回, 局_动作, 管理员Id, 管理员账号, constant.T提现操作人_管理员, ip, 原因)
}

func (j *S_RmbWithdraw) B标记已付款(数据库 *gorm.DB, id int, 管理员Id int, 管理员账号 string, ip string) error {
	var 局_提现单 dbm.DB_RmbWithdraw
	if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&局_提现单).Error; err != nil {
		return errors.New("提现单不存在")
	}
	if 局_提现单.Status != constant.T提现状态_付款中 {
		return errors.New("只有待付款提现可以标记已付款")
	}
	局_时间 := time.Now().Unix()
	局_错误 := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, constant.T提现状态_付款中).Updates(map[string]interface{}{
		"Status":       constant.T提现状态_已付款,
		"PayTime":      局_时间,
		"OperatorId":   管理员Id,
		"OperatorUser": 管理员账号,
		"UpdateTime":   局_时间,
	}).Error
	if 局_错误 == nil {
		局_错误 = j.提现_写日志(数据库, 局_提现单, constant.T提现状态_付款中, constant.T提现状态_已付款, constant.T提现动作_标记已付款, 管理员Id, 管理员账号, constant.T提现操作人_管理员, ip, "标记已付款")
	}
	return 局_错误
}

func (j *S_RmbWithdraw) R日志列表(数据库 *gorm.DB, 请求 T提现_列表请求) (int64, []dbm.DB_RmbWithdrawLog, error) {
	请求_规范列表参数(&请求)
	局_查询 := 数据库.Model(dbm.DB_RmbWithdrawLog{}).Order("Id DESC")
	if 请求.Uid > 0 {
		局_查询 = 局_查询.Where("Uid = ?", 请求.Uid)
	}
	if 请求.OrderNo != "" {
		局_查询 = 局_查询.Where("OrderNo = ?", 请求.OrderNo)
	}
	if 请求.Keywords != "" {
		局_查询 = 局_查询.Where("LOCATE(?, OrderNo)>0 OR LOCATE(?, OperatorUser)>0 OR LOCATE(?, Note)>0", 请求.Keywords, 请求.Keywords, 请求.Keywords)
	}
	var 局_数量 int64
	局_查询.Count(&局_数量)
	var 局_列表 []dbm.DB_RmbWithdrawLog
	局_错误 := 局_查询.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_列表).Error
	return 局_数量, 局_列表, 局_错误
}

func (j *S_RmbWithdraw) S删除(数据库 *gorm.DB, 请求 T提现_删除请求) (int64, error) {
	局_查询 := 数据库.Model(dbm.DB_RmbWithdraw{})
	switch 请求.Type {
	default:
		return 0, errors.New("Type错误")
	case 1:
		if len(请求.Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		局_查询 = 局_查询.Where("Id IN ?", 请求.Id)
	case 2:
		if strings.TrimSpace(请求.Keywords) == "" {
			return 0, errors.New("用户名不能为空")
		}
		局_查询 = 局_查询.Where("User = ?", strings.TrimSpace(请求.Keywords))
	case 3:
		局_查询 = 局_查询.Where("1 = 1")
	case 4:
		局_查询 = 局_查询.Where("CreateTime < ?", time.Now().Unix()-604800)
	case 5:
		局_查询 = 局_查询.Where("CreateTime < ?", time.Now().Unix()-2592000)
	case 6:
		局_查询 = 局_查询.Where("CreateTime < ?", time.Now().Unix()-7776000)
	case 8:
		局_查询 = 局_查询.Where("Status = ?", constant.T提现状态_已取消)
	}
	局_结果 := 局_查询.Delete(dbm.DB_RmbWithdraw{})
	return 局_结果.RowsAffected, 局_结果.Error
}

func (j *S_RmbWithdraw) C创建凭证令牌(id int, 管理员Id int, 管理员账号 string) (T提现_凭证令牌, error) {
	局_数据库 := global.Get局db()
	局_配置 := j.Q取配置(局_数据库)
	var 局_提现单 dbm.DB_RmbWithdraw
	if err := 局_数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, constant.T提现状态_付款中).First(&局_提现单).Error; err != nil {
		return T提现_凭证令牌{}, errors.New("提现单不存在或状态不允许上传凭证")
	}
	局_令牌 := 令牌_随机生成()
	局_信息 := T提现_凭证令牌{Token: 局_令牌, WithdrawId: id, AdminId: 管理员Id, AdminUser: 管理员账号, ExpireTime: time.Now().Unix() + 局_配置.VoucherTokenSeconds}
	global.H缓存.Set(集_提现凭证令牌前缀+局_令牌, 局_信息, time.Duration(局_配置.VoucherTokenSeconds)*time.Second)
	return 局_信息, nil
}

func (j *S_RmbWithdraw) Q取凭证令牌状态(token string) (T提现_凭证令牌, bool) {
	局_原始, 局_存在 := global.H缓存.Get(集_提现凭证令牌前缀 + token)
	if !局_存在 {
		return T提现_凭证令牌{}, false
	}
	局_信息, ok := 局_原始.(T提现_凭证令牌)
	if !ok || 局_信息.ExpireTime < time.Now().Unix() {
		return T提现_凭证令牌{}, false
	}
	return 局_信息, true
}

func (j *S_RmbWithdraw) A按令牌上传凭证(数据库 *gorm.DB, token string, 文件 *multipart.FileHeader, ip string) (string, error) {
	局_信息, 局_有效 := j.Q取凭证令牌状态(token)
	if !局_有效 {
		return "", errors.New("上传token无效或已过期")
	}
	if 局_信息.Used {
		return "", errors.New("上传token已使用")
	}
	局_路径, 局_错误 := j.S上传凭证(数据库, 局_信息.WithdrawId, 文件, 局_信息.AdminId, 局_信息.AdminUser, ip)
	if 局_错误 != nil {
		return "", 局_错误
	}
	局_信息.Used = true
	局_信息.UploadedPath = 局_路径
	global.H缓存.Set(集_提现凭证令牌前缀+token, 局_信息, time.Minute*5)
	return 局_路径, nil
}

func (j *S_RmbWithdraw) 提现_转退款(数据库 *gorm.DB, 提现单 dbm.DB_RmbWithdraw, 目标状态 int, 动作 int, 操作员Id int, 操作员账号 string, 操作员类型 int, ip string, 备注 string) error {
	局_时间 := time.Now().Unix()
	局_更新 := map[string]interface{}{
		"Status":       目标状态,
		"AdminReply":   备注,
		"OperatorId":   操作员Id,
		"OperatorUser": 操作员账号,
		"UpdateTime":   局_时间,
	}
	if 目标状态 == constant.T提现状态_已取消 {
		局_更新["CancelTime"] = 局_时间
	}
	局_结果 := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", 提现单.Id, 提现单.Status).Updates(局_更新)
	if 局_结果.Error != nil {
		return 局_结果.Error
	}
	if 局_结果.RowsAffected == 0 {
		return errors.New("状态已被其他人处理,请刷新")
	}
	if err := 数据库.Model(dbm.DB_User{}).Where("Id = ?", 提现单.Uid).Update("Rmb", gorm.Expr("Rmb + ?", 提现单.Amount)).Error; err != nil {
		return err
	}
	局_资金备注 := "提现驳回返还,提现单号:" + 提现单.OrderNo
	if 目标状态 == constant.T提现状态_已取消 {
		局_资金备注 = "提现取消返还,提现单号:" + 提现单.OrderNo
	} else if 动作 == constant.T提现动作_付款失败驳回 {
		局_资金备注 = "付款失败驳回返还,提现单号:" + 提现单.OrderNo
	}
	if err := j.提现_写资金日志(数据库, 提现单.Uid, 提现单.User, ip, 局_资金备注, 提现单.Amount); err != nil {
		return err
	}
	return j.提现_写日志(数据库, 提现单, 提现单.Status, 目标状态, 动作, 操作员Id, 操作员账号, 操作员类型, ip, 备注)
}

func (j *S_RmbWithdraw) 提现_写日志(数据库 *gorm.DB, 提现单 dbm.DB_RmbWithdraw, 前状态 int, 后状态 int, 动作 int, 操作员Id int, 操作员账号 string, 操作员类型 int, ip string, 备注 string) error {
	return 数据库.Model(dbm.DB_RmbWithdrawLog{}).Create(&dbm.DB_RmbWithdrawLog{
		WithdrawId:   提现单.Id,
		OrderNo:      提现单.OrderNo,
		Uid:          提现单.Uid,
		BeforeStatus: 前状态,
		AfterStatus:  后状态,
		Action:       动作,
		OperatorId:   操作员Id,
		OperatorUser: 操作员账号,
		OperatorType: 操作员类型,
		Ip:           ip,
		Note:         备注,
		Time:         time.Now().Unix(),
	}).Error
}

func (j *S_RmbWithdraw) 提现_写资金日志(数据库 *gorm.DB, uid int, user string, ip string, 备注 string, 金额 float64) error {
	var 局_新余额 float64
	_ = 数据库.Model(dbm.DB_User{}).Select("Rmb").Where("Id = ?", uid).Scan(&局_新余额).Error
	备注 = 备注 + "|新余额≈" + strconv.FormatFloat(局_新余额, 'f', 2, 64)
	return 数据库.Model(dbm.DB_LogMoney{}).Create(&dbm.DB_LogMoney{
		User:  user,
		Ip:    ip,
		Time:  time.Now().Unix(),
		Count: 金额,
		Note:  备注,
	}).Error
}

func (j *S_RmbWithdraw) 提现_统计金额(数据库 *gorm.DB, uid int, 状态数组 []int) float64 {
	var 局_合计 float64
	_ = 数据库.Model(dbm.DB_RmbWithdraw{}).Select("IFNULL(SUM(Amount), 0)").Where("Uid = ? AND Status IN ?", uid, 状态数组).Scan(&局_合计).Error
	return 局_合计
}

func (j *S_RmbWithdraw) 提现_检查间隔(数据库 *gorm.DB, uid int, 间隔 int64) error {
	if 间隔 <= 0 {
		return nil
	}
	var 局_最近一条 dbm.DB_RmbWithdraw
	if err := 数据库.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND Status NOT IN ?", uid, []int{constant.T提现状态_已驳回, constant.T提现状态_已取消}).Order("CreateTime DESC").First(&局_最近一条).Error; err == nil {
		if time.Now().Unix()-局_最近一条.CreateTime < 间隔 {
			return errors.New("未满足最小提现间隔")
		}
	}
	return nil
}

func (j *S_RmbWithdraw) 提现_风险标签(数据库 *gorm.DB, 提现单 dbm.DB_RmbWithdraw) []string {
	局_标签 := make([]string, 0)
	var 局_数量 int64
	数据库.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND Id <> ?", 提现单.Uid, 提现单.Id).Count(&局_数量)
	if 局_数量 == 0 {
		局_标签 = append(局_标签, "首次提现")
	}
	数据库.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND CreateTime >= ?", 提现单.Uid, time.Now().Unix()-86400).Count(&局_数量)
	if 局_数量 >= 2 {
		局_标签 = append(局_标签, "今日多次提现")
	}
	局_配置 := j.Q取配置(数据库)
	if 局_配置.MaxAmount > 0 && 提现单.Amount >= 局_配置.MaxAmount*0.9 {
		局_标签 = append(局_标签, "金额接近上限")
	}
	return 局_标签
}

func 图片_保存上传(文件 *multipart.FileHeader, 相对路径 string, 最大Mb int64) (string, error) {
	局_扩展名 := 文件_规范扩展名(文件.Filename)
	if 局_扩展名 != ".jpg" && 局_扩展名 != ".jpeg" && 局_扩展名 != ".png" {
		return "", errors.New("仅支持jpg/png/jpeg")
	}
	if 文件.Size > 最大Mb*1024*1024 {
		return "", errors.New("文件过大")
	}
	if _, err := 图片_解码(文件); err != nil {
		return "", err
	}
	局_绝对路径 := 路径_转绝对(相对路径)
	if err := os.MkdirAll(filepath.Dir(局_绝对路径), 0755); err != nil {
		return "", err
	}
	局_源, err := 文件.Open()
	if err != nil {
		return "", err
	}
	defer 局_源.Close()
	局_目标, err := os.Create(局_绝对路径)
	if err != nil {
		return "", err
	}
	defer 局_目标.Close()
	_, err = io.Copy(局_目标, 局_源)
	if err != nil {
		return "", err
	}
	return 相对路径, nil
}

func 图片_保存收款码(文件 *multipart.FileHeader, 相对路径 string, 最大Mb int64) (string, error) {
	if 文件.Size > 最大Mb*1024*1024 {
		return "", errors.New("尺寸错误")
	}
	局_图片, err := 图片_解码(文件)
	if err != nil {
		return "", err
	}
	局_定位点, err := 二维码_取定位点(局_图片)
	if err != nil {
		return "", errors.New("无法识别出图片二维码,请更换更清晰图片")
	}
	局_裁剪图 := 二维码_裁剪方形区域(局_图片, 局_定位点)
	局_缩放图 := 图片_最近邻缩放(局_裁剪图, 500, 500)
	局_绝对路径 := 路径_转绝对(相对路径)
	if err := os.MkdirAll(filepath.Dir(局_绝对路径), 0755); err != nil {
		return "", err
	}
	return 相对路径, 图片_写入文件(局_绝对路径, 局_缩放图, ".jpg")
}

func 图片_解码(文件 *multipart.FileHeader) (image.Image, error) {
	局_源, err := 文件.Open()
	if err != nil {
		return nil, err
	}
	defer 局_源.Close()
	局_图片, _, err := image.Decode(局_源)
	if err != nil {
		return nil, errors.New("invalid image file")
	}
	return 局_图片, nil
}

func 二维码_取定位点(图片 image.Image) ([]gozxing.ResultPoint, error) {
	局_位图, err := gozxing.NewBinaryBitmapFromImage(图片)
	if err != nil {
		return nil, err
	}
	局_结果, err := qrcode.NewQRCodeReader().Decode(局_位图, nil)
	if err != nil {
		return nil, err
	}
	if len(局_结果.GetText()) == 0 {
		return nil, errors.New("empty qrcode")
	}
	return 局_结果.GetResultPoints(), nil
}

func 二维码_裁剪方形区域(图片 image.Image, 定位点 []gozxing.ResultPoint) image.Image {
	局_边界 := 图片.Bounds()
	if len(定位点) == 0 {
		return 图片
	}
	局_最小X, 局_最小Y := math.MaxFloat64, math.MaxFloat64
	局_最大X, 局_最大Y := -math.MaxFloat64, -math.MaxFloat64
	for _, 局_点 := range 定位点 {
		局_X, 局_Y := float64(局_点.GetX()), float64(局_点.GetY())
		if 局_X < 局_最小X {
			局_最小X = 局_X
		}
		if 局_X > 局_最大X {
			局_最大X = 局_X
		}
		if 局_Y < 局_最小Y {
			局_最小Y = 局_Y
		}
		if 局_Y > 局_最大Y {
			局_最大Y = 局_Y
		}
	}
	局_宽 := 局_最大X - 局_最小X
	局_高 := 局_最大Y - 局_最小Y
	局_边长 := math.Max(局_宽, 局_高)
	if 局_边长 <= 0 {
		return 图片
	}
	局_边距 := math.Max(12, 局_边长*0.28)
	局_中心X := (局_最小X + 局_最大X) / 2
	局_中心Y := (局_最小Y + 局_最大Y) / 2
	局_边长 += 局_边距 * 2
	局_左 := int(math.Floor(局_中心X - 局_边长/2))
	局_上 := int(math.Floor(局_中心Y - 局_边长/2))
	局_右 := int(math.Ceil(局_中心X + 局_边长/2))
	局_下 := int(math.Ceil(局_中心Y + 局_边长/2))
	if 局_左 < 局_边界.Min.X {
		局_左 = 局_边界.Min.X
	}
	if 局_上 < 局_边界.Min.Y {
		局_上 = 局_边界.Min.Y
	}
	if 局_右 > 局_边界.Max.X {
		局_右 = 局_边界.Max.X
	}
	if 局_下 > 局_边界.Max.Y {
		局_下 = 局_边界.Max.Y
	}
	if 局_右 <= 局_左 || 局_下 <= 局_上 {
		return 图片
	}
	return 图片_复制区域(图片, image.Rect(局_左, 局_上, 局_右, 局_下))
}

func 图片_复制区域(源 image.Image, 区域 image.Rectangle) image.Image {
	局_目标 := image.NewRGBA(image.Rect(0, 0, 区域.Dx(), 区域.Dy()))
	for 局_Y := 0; 局_Y < 区域.Dy(); 局_Y++ {
		for 局_X := 0; 局_X < 区域.Dx(); 局_X++ {
			局_目标.Set(局_X, 局_Y, 源.At(区域.Min.X+局_X, 区域.Min.Y+局_Y))
		}
	}
	return 局_目标
}

func 图片_最近邻缩放(源 image.Image, 宽 int, 高 int) image.Image {
	局_边界 := 源.Bounds()
	局_目标 := image.NewRGBA(image.Rect(0, 0, 宽, 高))
	for 局_Y := 0; 局_Y < 高; 局_Y++ {
		局_源Y := 局_边界.Min.Y + 局_Y*局_边界.Dy()/高
		for 局_X := 0; 局_X < 宽; 局_X++ {
			局_源X := 局_边界.Min.X + 局_X*局_边界.Dx()/宽
			局_目标.Set(局_X, 局_Y, 源.At(局_源X, 局_源Y))
		}
	}
	return 局_目标
}

func 图片_写入文件(绝对路径 string, 图片 image.Image, 扩展名 string) error {
	局_目标, err := os.Create(绝对路径)
	if err != nil {
		return err
	}
	defer 局_目标.Close()
	if 扩展名 == ".png" {
		return png.Encode(局_目标, 图片)
	}
	return jpeg.Encode(局_目标, 图片, &jpeg.Options{Quality: 92})
}

func 凭证_日志路径标记(路径 string) string {
	return "voucherPath=" + 路径
}

func 凭证_从日志提取路径(备注 string) string {
	局_索引 := strings.Index(备注, "voucherPath=")
	if 局_索引 < 0 {
		return ""
	}
	局_路径 := 备注[局_索引+len("voucherPath="):]
	if 局_结束 := strings.IndexAny(局_路径, " \t\r\n"); 局_结束 >= 0 {
		局_路径 = 局_路径[:局_结束]
	}
	return 路径_规范图片路径(局_路径)
}

func 凭证_历史记录(日志列表 []dbm.DB_RmbWithdrawLog, 当前路径 string) []gin.H {
	局_历史 := make([]gin.H, 0)
	局_已见 := map[string]bool{}
	当前路径 = 路径_规范图片路径(当前路径)
	局_添加 := func(路径 string, 日志 dbm.DB_RmbWithdrawLog, 当前 bool) {
		路径 = 路径_规范图片路径(路径)
		if 路径 == "" || 局_已见[路径] {
			return
		}
		局_已见[路径] = true
		局_历史 = append(局_历史, gin.H{
			"path":         路径,
			"time":         日志.Time,
			"action":       日志.Action,
			"operatorUser": 日志.OperatorUser,
			"current":      当前,
		})
	}
	for 局_序号 := len(日志列表) - 1; 局_序号 >= 0; 局_序号-- {
		if 日志列表[局_序号].Action != constant.T提现动作_上传凭证 && 日志列表[局_序号].Action != constant.T提现动作_重新上传凭证 {
			continue
		}
		局_路径 := 凭证_从日志提取路径(日志列表[局_序号].Note)
		局_添加(局_路径, 日志列表[局_序号], 局_路径 == 当前路径)
	}
	局_添加(当前路径, dbm.DB_RmbWithdrawLog{Action: constant.T提现动作_上传凭证}, true)
	return 局_历史
}

func 文件_规范扩展名(名称 string) string {
	局_扩展名 := strings.ToLower(filepath.Ext(名称))
	if 局_扩展名 == ".jpeg" || 局_扩展名 == ".png" {
		return 局_扩展名
	}
	return ".jpg"
}

func 图片_收款码路径(uid int) string {
	return fmt.Sprintf("runtime/img/agent/payee_qr_%d.jpg", uid)
}

func 路径_转绝对(相对路径 string) string {
	局_相对 := strings.TrimPrefix(filepath.FromSlash(相对路径), string(filepath.Separator))
	return filepath.Join(global.GVA_CONFIG.Q取运行目录, 局_相对)
}

func 路径_规范图片路径(原始 string) string {
	局_路径 := strings.TrimSpace(原始)
	局_路径 = strings.TrimPrefix(局_路径, "/")
	局_路径 = filepath.ToSlash(局_路径)
	局_路径 = strings.TrimPrefix(局_路径, "./")
	if 局_路径 == "" || strings.Contains(局_路径, "..") || filepath.IsAbs(局_路径) {
		return ""
	}
	if !strings.HasPrefix(局_路径, "runtime/img/") {
		return ""
	}
	局_扩展名 := strings.ToLower(filepath.Ext(局_路径))
	if 局_扩展名 != ".jpg" && 局_扩展名 != ".jpeg" && 局_扩展名 != ".png" {
		return ""
	}
	return 局_路径
}

func 图片_取信息(相对路径 string) (T提现_图片信息, error) {
	局_绝对 := 路径_转绝对(相对路径)
	局_根 := filepath.Clean(filepath.Join(global.GVA_CONFIG.Q取运行目录, "runtime", "img"))
	局_干净绝对 := filepath.Clean(局_绝对)
	if 局_干净绝对 != 局_根 && !strings.HasPrefix(局_干净绝对, 局_根+string(filepath.Separator)) {
		return T提现_图片信息{}, errors.New("图片地址错误")
	}
	if _, ok := 文件_存在(局_干净绝对); !ok {
		return T提现_图片信息{}, errors.New("图片不存在")
	}
	return T提现_图片信息{AbsPath: 局_干净绝对, Ext: strings.ToLower(filepath.Ext(局_干净绝对))}, nil
}

func 文件_存在(路径 string) (os.FileInfo, bool) {
	局_信息, err := os.Stat(路径)
	return 局_信息, err == nil && !局_信息.IsDir()
}

func 文件_复制(源 string, 目标 string) error {
	if err := os.MkdirAll(filepath.Dir(目标), 0755); err != nil {
		return err
	}
	局_输入, err := os.Open(源)
	if err != nil {
		return err
	}
	defer 局_输入.Close()
	局_输出, err := os.Create(目标)
	if err != nil {
		return err
	}
	defer 局_输出.Close()
	_, err = io.Copy(局_输出, 局_输入)
	return err
}

func 请求_规范列表参数(请求 *T提现_列表请求) {
	if 请求.Page <= 0 {
		请求.Page = 1
	}
	if 请求.Size <= 0 {
		请求.Size = 10
	}
	if 请求.Size > 100 {
		请求.Size = 100
	}
}

func 单号_生成(uid int) string {
	return fmt.Sprintf("WD%d%d", time.Now().UnixNano(), uid)
}

func 令牌_随机生成() string {
	局_缓冲 := make([]byte, 24)
	_, _ = rand.Read(局_缓冲)
	局_哈希 := sha256.Sum256([]byte(fmt.Sprintf("%x-%d", 局_缓冲, time.Now().UnixNano())))
	return hex.EncodeToString(局_哈希[:])
}
