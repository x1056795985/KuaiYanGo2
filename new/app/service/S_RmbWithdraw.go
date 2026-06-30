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
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"gorm.io/gorm"

	"server/global"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
)

const (
	WithdrawStatusPending  = 1
	WithdrawStatusRejected = 2
	WithdrawStatusPaying   = 3
	WithdrawStatusPaid     = 4
	WithdrawStatusCanceled = 5
)

const (
	WithdrawActionCreate          = 1
	WithdrawActionUserCancel      = 2
	WithdrawActionAuditPass       = 3
	WithdrawActionReject          = 4
	WithdrawActionUploadVoucher   = 5
	WithdrawActionReuploadVoucher = 6
	WithdrawActionMarkPaid        = 7
	WithdrawActionPayFailReject   = 8
)

const (
	WithdrawOperatorUser  = 1
	WithdrawOperatorAdmin = 2
	WithdrawOperatorSys   = 3
)

const (
	withdrawConfigKey     = "agentWithdrawConfig"
	withdrawVoucherPrefix = "withdrawVoucherToken:"
)

type WithdrawConfig struct {
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

type WithdrawCreateRequest struct {
	Amount         float64 `json:"amount"`
	PayeeType      int     `json:"payeeType"`
	UseLastPayeeQr bool    `json:"useLastPayeeQr"`
	PayeeAccount   string  `json:"payeeAccount"`
	PayeeName      string  `json:"payeeName"`
	UserNote       string  `json:"userNote"`
	RequestId      string  `json:"requestId"`
}

type WithdrawListRequest struct {
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

type WithdrawDeleteRequest struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

type WithdrawImageInfo struct {
	AbsPath string
	Ext     string
}

type WithdrawVoucherToken struct {
	Token        string `json:"token"`
	WithdrawId   int    `json:"withdrawId"`
	AdminId      int    `json:"adminId"`
	AdminUser    string `json:"adminUser"`
	ExpireTime   int64  `json:"expireTime"`
	Used         bool   `json:"used"`
	UploadedPath string `json:"uploadedPath"`
}

type S_RmbWithdraw struct{}

func DefaultWithdrawConfig() WithdrawConfig {
	return WithdrawConfig{
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

func (s *S_RmbWithdraw) GetConfig(tx *gorm.DB) WithdrawConfig {
	cfg := DefaultWithdrawConfig()
	var setting dbm.DB_Setting
	if err := tx.Model(dbm.DB_Setting{}).Where("ItemKey = ?", withdrawConfigKey).First(&setting).Error; err == nil && setting.ItemValue != "" {
		_ = json.Unmarshal([]byte(setting.ItemValue), &cfg)
	}
	if cfg.VoucherTokenSeconds <= 0 {
		cfg.VoucherTokenSeconds = 300
	}
	if cfg.PayeeQrMaxSizeMb <= 0 {
		cfg.PayeeQrMaxSizeMb = 5
	}
	if cfg.VoucherMaxSizeMb <= 0 {
		cfg.VoucherMaxSizeMb = 10
	}
	return cfg
}

func (s *S_RmbWithdraw) SaveConfig(tx *gorm.DB, cfg WithdrawConfig) error {
	if cfg.MinAmount < 0 || cfg.MaxAmount < 0 || (cfg.MaxAmount > 0 && cfg.MaxAmount < cfg.MinAmount) {
		return errors.New("提现金额配置不正确")
	}
	data, _ := json.Marshal(cfg)
	var setting dbm.DB_Setting
	if err := tx.Model(dbm.DB_Setting{}).Where("ItemKey = ?", withdrawConfigKey).First(&setting).Error; err == nil {
		return tx.Model(dbm.DB_Setting{}).Where("ItemKey = ?", withdrawConfigKey).Update("ItemValue", string(data)).Error
	}
	return tx.Model(dbm.DB_Setting{}).Create(&dbm.DB_Setting{ItemKey: withdrawConfigKey, ItemValue: string(data)}).Error
}

func (s *S_RmbWithdraw) GetAgentConfig(tx *gorm.DB, uid int) (gin.H, error) {
	cfg := s.GetConfig(tx)
	var user DB.DB_User
	if err := tx.Model(DB.DB_User{}).Where("Id = ?", uid).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	frozen := s.sumAmount(tx, uid, []int{WithdrawStatusPending, WithdrawStatusPaying})
	auditing := s.sumAmount(tx, uid, []int{WithdrawStatusPending})
	var last dbm.DB_RmbWithdraw
	_ = tx.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ?", uid).Order("Id DESC").First(&last).Error

	next := int64(0)
	if cfg.IntervalSeconds > 0 {
		var lastValid dbm.DB_RmbWithdraw
		if err := tx.Model(dbm.DB_RmbWithdraw{}).
			Where("Uid = ? AND Status NOT IN ?", uid, []int{WithdrawStatusRejected, WithdrawStatusCanceled}).
			Order("CreateTime DESC").First(&lastValid).Error; err == nil {
			next = lastValid.CreateTime + cfg.IntervalSeconds
			if next < time.Now().Unix() {
				next = 0
			}
		}
	}

	qr := payeeQrPath(uid)
	_, hasQr := fileExists(absPath(qr))
	return gin.H{
		"enable":             cfg.Enable,
		"minAmount":          cfg.MinAmount,
		"maxAmount":          cfg.MaxAmount,
		"intervalSeconds":    cfg.IntervalSeconds,
		"allowUserCancel":    cfg.AllowUserCancel,
		"requirePayeeQr":     cfg.RequirePayeeQr,
		"allowPayeeAccount":  cfg.AllowPayeeAccount,
		"nextWithdrawTime":   next,
		"availableRmb":       user.Rmb,
		"frozenAmount":       frozen,
		"auditingAmount":     auditing,
		"lastAmount":         last.Amount,
		"lastWithdrawAmount": last.Amount,
		"hasPayeeQr":         hasQr,
		"payeeQrPath":        qr,
	}, nil
}

func (s *S_RmbWithdraw) UploadPayeeQr(uid int, file *multipart.FileHeader) (string, error) {
	cfg := s.GetConfig(global.GVA_DB)
	return savePayeeQrImage(file, payeeQrPath(uid), cfg.PayeeQrMaxSizeMb)
}

func (s *S_RmbWithdraw) GetAgentImage(tx *gorm.DB, uid int, path string) (WithdrawImageInfo, error) {
	path = normalizeRuntimeImagePath(path)
	if path == "" {
		return WithdrawImageInfo{}, errors.New("图片地址错误")
	}
	if path == payeeQrPath(uid) {
		return imageInfo(path)
	}

	var count int64
	err := tx.Model(dbm.DB_RmbWithdraw{}).
		Where("Uid = ? AND (PayeeQrPath = ? OR VoucherPath = ?)", uid, path, path).
		Count(&count).Error
	if err != nil {
		return WithdrawImageInfo{}, err
	}
	if count == 0 {
		return WithdrawImageInfo{}, errors.New("无权查看该图片")
	}
	return imageInfo(path)
}

func (s *S_RmbWithdraw) GetAdminImage(tx *gorm.DB, path string) (WithdrawImageInfo, error) {
	path = normalizeRuntimeImagePath(path)
	if path == "" {
		return WithdrawImageInfo{}, errors.New("图片地址错误")
	}

	var count int64
	err := tx.Model(dbm.DB_RmbWithdraw{}).
		Where("PayeeQrPath = ? OR VoucherPath = ? OR Id IN (?)", path, path,
			tx.Model(dbm.DB_RmbWithdrawLog{}).Select("WithdrawId").Where("Action IN ? AND LOCATE(?, Note)>0", []int{WithdrawActionUploadVoucher, WithdrawActionReuploadVoucher}, voucherLogPathToken(path))).
		Count(&count).Error
	if err != nil {
		return WithdrawImageInfo{}, err
	}
	if count == 0 {
		return WithdrawImageInfo{}, errors.New("无权查看该图片")
	}
	return imageInfo(path)
}

func (s *S_RmbWithdraw) UploadVoucher(withdrawId int, file *multipart.FileHeader, operatorId int, operatorUser string, ip string) (string, error) {
	path := fmt.Sprintf("runtime/img/admin/withdraw_voucher_%d_%d%s", withdrawId, time.Now().Unix(), normalizedExt(file.Filename))
	cfg := s.GetConfig(global.GVA_DB)
	path, err := saveUploadedImage(file, path, cfg.VoucherMaxSizeMb)
	if err != nil {
		return "", err
	}

	tx := global.GVA_DB.Begin()
	var withdraw dbm.DB_RmbWithdraw
	if err = tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", withdrawId).First(&withdraw).Error; err != nil {
		tx.Rollback()
		return "", errors.New("提现单不存在")
	}
	if withdraw.Status != WithdrawStatusPaying && withdraw.Status != WithdrawStatusPaid {
		tx.Rollback()
		return "", errors.New("当前状态不允许上传凭证")
	}

	action := WithdrawActionUploadVoucher
	if withdraw.VoucherPath != "" {
		action = WithdrawActionReuploadVoucher
	}
	err = tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", withdrawId).Updates(map[string]interface{}{
		"VoucherPath":  path,
		"OperatorId":   operatorId,
		"OperatorUser": operatorUser,
		"UpdateTime":   time.Now().Unix(),
	}).Error
	if err == nil {
		err = s.writeLog(tx, withdraw, withdraw.Status, withdraw.Status, action, operatorId, operatorUser, WithdrawOperatorAdmin, ip, "upload voucher "+voucherLogPathToken(path))
	}
	return path, finishTx(tx, err)
}

func (s *S_RmbWithdraw) Create(uid int, user string, ip string, req WithdrawCreateRequest) (dbm.DB_RmbWithdraw, error) {
	var empty dbm.DB_RmbWithdraw
	if req.Amount <= 0 {
		return empty, errors.New("提现金额必须大于0")
	}

	tx := global.GVA_DB.Begin()
	cfg := s.GetConfig(tx)
	if !cfg.Enable {
		tx.Rollback()
		return empty, errors.New("代理提现未启用")
	}
	if req.Amount < cfg.MinAmount || (cfg.MaxAmount > 0 && req.Amount > cfg.MaxAmount) {
		tx.Rollback()
		return empty, errors.New("提现金额不符合规则")
	}
	if req.RequestId != "" {
		var old dbm.DB_RmbWithdraw
		if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND RequestId = ?", uid, req.RequestId).First(&old).Error; err == nil {
			tx.Rollback()
			return old, nil
		}
	}

	var dbUser DB.DB_User
	if err := tx.Model(DB.DB_User{}).Where("Id = ?", uid).First(&dbUser).Error; err != nil {
		tx.Rollback()
		return empty, errors.New("用户不存在")
	}
	if dbUser.Status != 1 {
		tx.Rollback()
		return empty, errors.New("用户状态不正常")
	}
	if dbUser.AgentDiscount <= 0 && dbUser.UPAgentId == 0 {
		tx.Rollback()
		return empty, errors.New("当前账号不是代理")
	}
	if err := s.checkInterval(tx, uid, cfg.IntervalSeconds); err != nil {
		tx.Rollback()
		return empty, err
	}

	payeeQr := ""
	if req.PayeeType == 1 || cfg.RequirePayeeQr {
		src := payeeQrPath(uid)
		if _, ok := fileExists(absPath(src)); !ok {
			tx.Rollback()
			return empty, errors.New("请先上传收款码")
		}
		payeeQr = fmt.Sprintf("runtime/img/agent/withdraw/payee_qr_%d_%d.jpg", uid, time.Now().Unix())
		if err := copyFile(absPath(src), absPath(payeeQr)); err != nil {
			tx.Rollback()
			return empty, err
		}
	} else if req.PayeeType == 2 {
		if !cfg.AllowPayeeAccount {
			tx.Rollback()
			return empty, errors.New("当前不允许填写收款账号")
		}
		if strings.TrimSpace(req.PayeeAccount) == "" {
			tx.Rollback()
			return empty, errors.New("收款账号不能为空")
		}
	} else {
		tx.Rollback()
		return empty, errors.New("收款方式错误")
	}

	ret := tx.Model(DB.DB_User{}).Where("Id = ? AND Rmb >= ?", uid, req.Amount).Update("Rmb", gorm.Expr("Rmb - ?", req.Amount))
	if ret.Error != nil {
		tx.Rollback()
		return empty, ret.Error
	}
	if ret.RowsAffected == 0 {
		tx.Rollback()
		return empty, errors.New("余额不足")
	}

	now := time.Now().Unix()
	raw, _ := json.Marshal(gin.H{"payeeType": req.PayeeType, "payeeQrPath": payeeQr, "payeeAccount": req.PayeeAccount, "payeeName": req.PayeeName})
	withdraw := dbm.DB_RmbWithdraw{
		OrderNo:      makeOrderNo(uid),
		RequestId:    req.RequestId,
		Uid:          uid,
		User:         user,
		WithdrawType: 1,
		Amount:       req.Amount,
		Status:       WithdrawStatusPending,
		UserNote:     req.UserNote,
		PayeeType:    req.PayeeType,
		PayeeQrPath:  payeeQr,
		PayeeAccount: req.PayeeAccount,
		PayeeName:    req.PayeeName,
		PayeeRawInfo: string(raw),
		CreateTime:   now,
		UpdateTime:   now,
		Ip:           ip,
	}
	err := tx.Model(dbm.DB_RmbWithdraw{}).Create(&withdraw).Error
	if err == nil {
		err = s.writeMoneyLog(tx, uid, user, ip, "提现冻结,提现单号:"+withdraw.OrderNo, -req.Amount)
	}
	if err == nil {
		err = s.writeLog(tx, withdraw, 0, WithdrawStatusPending, WithdrawActionCreate, uid, user, WithdrawOperatorUser, ip, "提交提现申请")
	}
	return withdraw, finishTx(tx, err)
}

func (s *S_RmbWithdraw) List(tx *gorm.DB, req WithdrawListRequest, uid int) (int64, []dbm.DB_RmbWithdraw, error) {
	normalizeListRequest(&req)
	db := tx.Model(dbm.DB_RmbWithdraw{}).Order("Id DESC")
	if uid > 0 {
		db = db.Where("Uid = ?", uid)
	}
	if req.Status > 0 {
		db = db.Where("Status = ?", req.Status)
	}
	if req.Uid > 0 {
		db = db.Where("Uid = ?", req.Uid)
	}
	if req.User != "" {
		db = db.Where("User = ?", req.User)
	}
	if req.OrderNo != "" {
		db = db.Where("OrderNo = ?", req.OrderNo)
	}
	if req.MinAmount > 0 {
		db = db.Where("Amount >= ?", req.MinAmount)
	}
	if req.MaxAmount > 0 {
		db = db.Where("Amount <= ?", req.MaxAmount)
	}
	if len(req.RegisterTime) == 2 && req.RegisterTime[0] != "" && req.RegisterTime[1] != "" {
		start, _ := strconv.ParseInt(req.RegisterTime[0], 10, 64)
		end, _ := strconv.ParseInt(req.RegisterTime[1], 10, 64)
		db = db.Where("CreateTime >= ? AND CreateTime < ?", start, end+86400)
	}
	if req.Keywords != "" {
		switch req.Type {
		case 1:
			db = db.Where("User = ?", req.Keywords)
		case 2:
			db = db.Where("Uid = ?", req.Keywords)
		case 3:
			db = db.Where("OrderNo = ?", req.Keywords)
		case 4:
			db = db.Where("Amount = ?", req.Keywords)
		default:
			db = db.Where("LOCATE(?, OrderNo)>0 OR LOCATE(?, User)>0", req.Keywords, req.Keywords)
		}
	}

	var count int64
	if req.Count > 500000 {
		count = req.Count
	} else {
		db.Count(&count)
	}
	var list []dbm.DB_RmbWithdraw
	err := db.Limit(req.Size).Offset((req.Page - 1) * req.Size).Find(&list).Error
	return count, list, err
}

func (s *S_RmbWithdraw) Detail(tx *gorm.DB, id int, uid int) (gin.H, error) {
	var withdraw dbm.DB_RmbWithdraw
	db := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id)
	if uid > 0 {
		db = db.Where("Uid = ?", uid)
	}
	if err := db.First(&withdraw).Error; err != nil {
		return nil, errors.New("提现单不存在")
	}

	var logs []dbm.DB_RmbWithdrawLog
	_ = tx.Model(dbm.DB_RmbWithdrawLog{}).Where("WithdrawId = ?", id).Order("Id ASC").Find(&logs).Error
	var user DB.DB_User
	_ = tx.Model(DB.DB_User{}).Where("Id = ?", withdraw.Uid).First(&user).Error
	return gin.H{"info": withdraw, "logs": logs, "user": user, "riskTags": s.riskTags(tx, withdraw), "voucherHistory": voucherHistoryFromLogs(logs, withdraw.VoucherPath)}, nil
}

func (s *S_RmbWithdraw) Cancel(id int, uid int, user string, ip string) error {
	tx := global.GVA_DB.Begin()
	var withdraw dbm.DB_RmbWithdraw
	if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Uid = ?", id, uid).First(&withdraw).Error; err != nil {
		tx.Rollback()
		return errors.New("提现单不存在")
	}
	if withdraw.Status != WithdrawStatusPending {
		tx.Rollback()
		return errors.New("只有待审核提现可以取消")
	}
	if !s.GetConfig(tx).AllowUserCancel {
		tx.Rollback()
		return errors.New("当前不允许用户取消提现")
	}
	err := s.changeToRefund(tx, withdraw, WithdrawStatusCanceled, WithdrawActionUserCancel, uid, user, WithdrawOperatorUser, ip, "用户取消提现")
	return finishTx(tx, err)
}

func (s *S_RmbWithdraw) AuditPass(id int, adminId int, adminUser string, ip string) error {
	tx := global.GVA_DB.Begin()
	var withdraw dbm.DB_RmbWithdraw
	if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&withdraw).Error; err != nil {
		tx.Rollback()
		return errors.New("提现单不存在")
	}
	if withdraw.Status != WithdrawStatusPending {
		tx.Rollback()
		return errors.New("只有待审核提现可以审核通过")
	}
	now := time.Now().Unix()
	err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, WithdrawStatusPending).Updates(map[string]interface{}{
		"Status":       WithdrawStatusPaying,
		"AuditTime":    now,
		"OperatorId":   adminId,
		"OperatorUser": adminUser,
		"UpdateTime":   now,
	}).Error
	if err == nil {
		err = s.writeLog(tx, withdraw, WithdrawStatusPending, WithdrawStatusPaying, WithdrawActionAuditPass, adminId, adminUser, WithdrawOperatorAdmin, ip, "审核通过")
	}
	return finishTx(tx, err)
}

func (s *S_RmbWithdraw) Reject(id int, reason string, adminId int, adminUser string, ip string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("驳回原因不能为空")
	}
	tx := global.GVA_DB.Begin()
	var withdraw dbm.DB_RmbWithdraw
	if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&withdraw).Error; err != nil {
		tx.Rollback()
		return errors.New("提现单不存在")
	}
	if withdraw.Status != WithdrawStatusPending && withdraw.Status != WithdrawStatusPaying {
		tx.Rollback()
		return errors.New("当前状态不允许驳回")
	}
	action := WithdrawActionReject
	if withdraw.Status == WithdrawStatusPaying {
		action = WithdrawActionPayFailReject
	}
	err := s.changeToRefund(tx, withdraw, WithdrawStatusRejected, action, adminId, adminUser, WithdrawOperatorAdmin, ip, reason)
	return finishTx(tx, err)
}

func (s *S_RmbWithdraw) MarkPaid(id int, adminId int, adminUser string, ip string) error {
	tx := global.GVA_DB.Begin()
	var withdraw dbm.DB_RmbWithdraw
	if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ?", id).First(&withdraw).Error; err != nil {
		tx.Rollback()
		return errors.New("提现单不存在")
	}
	if withdraw.Status != WithdrawStatusPaying {
		tx.Rollback()
		return errors.New("只有待付款提现可以标记已付款")
	}
	now := time.Now().Unix()
	err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, WithdrawStatusPaying).Updates(map[string]interface{}{
		"Status":       WithdrawStatusPaid,
		"PayTime":      now,
		"OperatorId":   adminId,
		"OperatorUser": adminUser,
		"UpdateTime":   now,
	}).Error
	if err == nil {
		err = s.writeLog(tx, withdraw, WithdrawStatusPaying, WithdrawStatusPaid, WithdrawActionMarkPaid, adminId, adminUser, WithdrawOperatorAdmin, ip, "标记已付款")
	}
	return finishTx(tx, err)
}

func (s *S_RmbWithdraw) Logs(tx *gorm.DB, req WithdrawListRequest) (int64, []dbm.DB_RmbWithdrawLog, error) {
	normalizeListRequest(&req)
	db := tx.Model(dbm.DB_RmbWithdrawLog{}).Order("Id DESC")
	if req.Uid > 0 {
		db = db.Where("Uid = ?", req.Uid)
	}
	if req.OrderNo != "" {
		db = db.Where("OrderNo = ?", req.OrderNo)
	}
	if req.Keywords != "" {
		db = db.Where("LOCATE(?, OrderNo)>0 OR LOCATE(?, OperatorUser)>0 OR LOCATE(?, Note)>0", req.Keywords, req.Keywords, req.Keywords)
	}
	var count int64
	db.Count(&count)
	var list []dbm.DB_RmbWithdrawLog
	err := db.Limit(req.Size).Offset((req.Page - 1) * req.Size).Find(&list).Error
	return count, list, err
}

func (s *S_RmbWithdraw) Delete(tx *gorm.DB, req WithdrawDeleteRequest) (int64, error) {
	db := tx.Model(dbm.DB_RmbWithdraw{})
	switch req.Type {
	default:
		return 0, errors.New("Type错误")
	case 1:
		if len(req.Id) == 0 {
			return 0, errors.New("Id数组没有要删除的ID")
		}
		db = db.Where("Id IN ?", req.Id)
	case 2:
		if strings.TrimSpace(req.Keywords) == "" {
			return 0, errors.New("用户名不能为空")
		}
		db = db.Where("User = ?", strings.TrimSpace(req.Keywords))
	case 3:
		db = db.Where("1 = 1")
	case 4:
		db = db.Where("CreateTime < ?", time.Now().Unix()-604800)
	case 5:
		db = db.Where("CreateTime < ?", time.Now().Unix()-2592000)
	case 6:
		db = db.Where("CreateTime < ?", time.Now().Unix()-7776000)
	case 8:
		db = db.Where("Status = ?", WithdrawStatusCanceled)
	}
	result := db.Delete(dbm.DB_RmbWithdraw{})
	return result.RowsAffected, result.Error
}

func (s *S_RmbWithdraw) CreateVoucherToken(id int, adminId int, adminUser string) (WithdrawVoucherToken, error) {
	cfg := s.GetConfig(global.GVA_DB)
	var withdraw dbm.DB_RmbWithdraw
	if err := global.GVA_DB.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", id, WithdrawStatusPaying).First(&withdraw).Error; err != nil {
		return WithdrawVoucherToken{}, errors.New("提现单不存在或状态不允许上传凭证")
	}
	token := randomToken()
	info := WithdrawVoucherToken{Token: token, WithdrawId: id, AdminId: adminId, AdminUser: adminUser, ExpireTime: time.Now().Unix() + cfg.VoucherTokenSeconds}
	global.H缓存.Set(withdrawVoucherPrefix+token, info, time.Duration(cfg.VoucherTokenSeconds)*time.Second)
	return info, nil
}

func (s *S_RmbWithdraw) GetVoucherTokenStatus(token string) (WithdrawVoucherToken, bool) {
	raw, ok := global.H缓存.Get(withdrawVoucherPrefix + token)
	if !ok {
		return WithdrawVoucherToken{}, false
	}
	info, ok := raw.(WithdrawVoucherToken)
	if !ok || info.ExpireTime < time.Now().Unix() {
		return WithdrawVoucherToken{}, false
	}
	return info, true
}

func (s *S_RmbWithdraw) UploadVoucherByToken(token string, file *multipart.FileHeader, ip string) (string, error) {
	info, ok := s.GetVoucherTokenStatus(token)
	if !ok {
		return "", errors.New("上传token无效或已过期")
	}
	if info.Used {
		return "", errors.New("上传token已使用")
	}
	path, err := s.UploadVoucher(info.WithdrawId, file, info.AdminId, info.AdminUser, ip)
	if err != nil {
		return "", err
	}
	info.Used = true
	info.UploadedPath = path
	global.H缓存.Set(withdrawVoucherPrefix+token, info, time.Minute*5)
	return path, nil
}

func (s *S_RmbWithdraw) changeToRefund(tx *gorm.DB, w dbm.DB_RmbWithdraw, afterStatus int, action int, operatorId int, operatorUser string, operatorType int, ip string, note string) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"Status":       afterStatus,
		"AdminReply":   note,
		"OperatorId":   operatorId,
		"OperatorUser": operatorUser,
		"UpdateTime":   now,
	}
	if afterStatus == WithdrawStatusCanceled {
		updates["CancelTime"] = now
	}
	ret := tx.Model(dbm.DB_RmbWithdraw{}).Where("Id = ? AND Status = ?", w.Id, w.Status).Updates(updates)
	if ret.Error != nil {
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		return errors.New("状态已被其他人处理,请刷新")
	}
	if err := tx.Model(DB.DB_User{}).Where("Id = ?", w.Uid).Update("Rmb", gorm.Expr("Rmb + ?", w.Amount)).Error; err != nil {
		return err
	}
	moneyNote := "提现驳回返还,提现单号:" + w.OrderNo
	if afterStatus == WithdrawStatusCanceled {
		moneyNote = "提现取消返还,提现单号:" + w.OrderNo
	} else if action == WithdrawActionPayFailReject {
		moneyNote = "付款失败驳回返还,提现单号:" + w.OrderNo
	}
	if err := s.writeMoneyLog(tx, w.Uid, w.User, ip, moneyNote, w.Amount); err != nil {
		return err
	}
	return s.writeLog(tx, w, w.Status, afterStatus, action, operatorId, operatorUser, operatorType, ip, note)
}

func (s *S_RmbWithdraw) writeLog(tx *gorm.DB, w dbm.DB_RmbWithdraw, beforeStatus int, afterStatus int, action int, operatorId int, operatorUser string, operatorType int, ip string, note string) error {
	return tx.Model(dbm.DB_RmbWithdrawLog{}).Create(&dbm.DB_RmbWithdrawLog{
		WithdrawId:   w.Id,
		OrderNo:      w.OrderNo,
		Uid:          w.Uid,
		BeforeStatus: beforeStatus,
		AfterStatus:  afterStatus,
		Action:       action,
		OperatorId:   operatorId,
		OperatorUser: operatorUser,
		OperatorType: operatorType,
		Ip:           ip,
		Note:         note,
		Time:         time.Now().Unix(),
	}).Error
}

func (s *S_RmbWithdraw) writeMoneyLog(tx *gorm.DB, uid int, user string, ip string, note string, amount float64) error {
	var 局_新余额 float64
	_ = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", uid).Scan(&局_新余额).Error
	note = note + "|新余额≈" + strconv.FormatFloat(局_新余额, 'f', 2, 64)
	return tx.Model(DB.DB_LogMoney{}).Create(&DB.DB_LogMoney{
		User:  user,
		Ip:    ip,
		Time:  time.Now().Unix(),
		Count: amount,
		Note:  note,
	}).Error
}

func (s *S_RmbWithdraw) sumAmount(tx *gorm.DB, uid int, status []int) float64 {
	var sum float64
	_ = tx.Model(dbm.DB_RmbWithdraw{}).Select("IFNULL(SUM(Amount), 0)").Where("Uid = ? AND Status IN ?", uid, status).Scan(&sum).Error
	return sum
}

func (s *S_RmbWithdraw) checkInterval(tx *gorm.DB, uid int, interval int64) error {
	if interval <= 0 {
		return nil
	}
	var last dbm.DB_RmbWithdraw
	if err := tx.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND Status NOT IN ?", uid, []int{WithdrawStatusRejected, WithdrawStatusCanceled}).Order("CreateTime DESC").First(&last).Error; err == nil {
		if time.Now().Unix()-last.CreateTime < interval {
			return errors.New("未满足最小提现间隔")
		}
	}
	return nil
}

func (s *S_RmbWithdraw) riskTags(tx *gorm.DB, w dbm.DB_RmbWithdraw) []string {
	tags := make([]string, 0)
	var count int64
	tx.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND Id <> ?", w.Uid, w.Id).Count(&count)
	if count == 0 {
		tags = append(tags, "首次提现")
	}
	tx.Model(dbm.DB_RmbWithdraw{}).Where("Uid = ? AND CreateTime >= ?", w.Uid, time.Now().Unix()-86400).Count(&count)
	if count >= 2 {
		tags = append(tags, "今日多次提现")
	}
	cfg := s.GetConfig(tx)
	if cfg.MaxAmount > 0 && w.Amount >= cfg.MaxAmount*0.9 {
		tags = append(tags, "金额接近上限")
	}
	return tags
}

func saveUploadedImage(file *multipart.FileHeader, relPath string, maxMb int64) (string, error) {
	ext := normalizedExt(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("仅支持jpg/png/jpeg")
	}
	if file.Size > maxMb*1024*1024 {
		return "", errors.New("文件过大")
	}
	if _, err := decodeUploadedImage(file); err != nil {
		return "", err
	}
	abs := absPath(relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return "", err
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(abs)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}
	return relPath, nil
}

func savePayeeQrImage(file *multipart.FileHeader, relPath string, maxMb int64) (string, error) {
	if file.Size > maxMb*1024*1024 {
		return "", errors.New("尺寸错误")
	}
	img, err := decodeUploadedImage(file)
	if err != nil {
		return "", err
	}
	points, err := decodeQrPoints(img)
	if err != nil {
		return "", errors.New("无法识别出图片二维码,请更换更清晰图片")
	}
	cropped := cropQrSquare(img, points)
	resized := resizeNearest(cropped, 500, 500)
	abs := absPath(relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return "", err
	}
	return relPath, writeImage(abs, resized, ".jpg")
}

func decodeUploadedImage(file *multipart.FileHeader) (image.Image, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.New("invalid image file")
	}
	return img, nil
}

func decodeQrPoints(img image.Image) ([]gozxing.ResultPoint, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}
	result, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return nil, err
	}
	if len(result.GetText()) == 0 {
		return nil, errors.New("empty qrcode")
	}
	return result.GetResultPoints(), nil
}

func cropQrSquare(img image.Image, points []gozxing.ResultPoint) image.Image {
	bounds := img.Bounds()
	if len(points) == 0 {
		return img
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, point := range points {
		x, y := float64(point.GetX()), float64(point.GetY())
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	width := maxX - minX
	height := maxY - minY
	side := math.Max(width, height)
	if side <= 0 {
		return img
	}
	margin := math.Max(12, side*0.28)
	centerX := (minX + maxX) / 2
	centerY := (minY + maxY) / 2
	side += margin * 2
	left := int(math.Floor(centerX - side/2))
	top := int(math.Floor(centerY - side/2))
	right := int(math.Ceil(centerX + side/2))
	bottom := int(math.Ceil(centerY + side/2))
	if left < bounds.Min.X {
		left = bounds.Min.X
	}
	if top < bounds.Min.Y {
		top = bounds.Min.Y
	}
	if right > bounds.Max.X {
		right = bounds.Max.X
	}
	if bottom > bounds.Max.Y {
		bottom = bounds.Max.Y
	}
	if right <= left || bottom <= top {
		return img
	}
	return copyImageRegion(img, image.Rect(left, top, right, bottom))
}

func copyImageRegion(src image.Image, rect image.Rectangle) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	for y := 0; y < rect.Dy(); y++ {
		for x := 0; x < rect.Dx(); x++ {
			dst.Set(x, y, src.At(rect.Min.X+x, rect.Min.Y+y))
		}
	}
	return dst
}

func resizeNearest(src image.Image, width int, height int) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		sy := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < width; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func writeImage(abs string, img image.Image, ext string) error {
	dst, err := os.Create(abs)
	if err != nil {
		return err
	}
	defer dst.Close()
	if ext == ".png" {
		return png.Encode(dst, img)
	}
	return jpeg.Encode(dst, img, &jpeg.Options{Quality: 92})
}

func voucherLogPathToken(path string) string {
	return "voucherPath=" + path
}

func extractVoucherLogPath(note string) string {
	idx := strings.Index(note, "voucherPath=")
	if idx < 0 {
		return ""
	}
	path := note[idx+len("voucherPath="):]
	if end := strings.IndexAny(path, " \t\r\n"); end >= 0 {
		path = path[:end]
	}
	return normalizeRuntimeImagePath(path)
}

func voucherHistoryFromLogs(logs []dbm.DB_RmbWithdrawLog, currentPath string) []gin.H {
	history := make([]gin.H, 0)
	seen := map[string]bool{}
	currentPath = normalizeRuntimeImagePath(currentPath)
	add := func(path string, log dbm.DB_RmbWithdrawLog, current bool) {
		path = normalizeRuntimeImagePath(path)
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		history = append(history, gin.H{
			"path":         path,
			"time":         log.Time,
			"action":       log.Action,
			"operatorUser": log.OperatorUser,
			"current":      current,
		})
	}
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].Action != WithdrawActionUploadVoucher && logs[i].Action != WithdrawActionReuploadVoucher {
			continue
		}
		path := extractVoucherLogPath(logs[i].Note)
		add(path, logs[i], path == currentPath)
	}
	add(currentPath, dbm.DB_RmbWithdrawLog{Action: WithdrawActionUploadVoucher}, true)
	return history
}

func normalizedExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".jpeg" || ext == ".png" {
		return ext
	}
	return ".jpg"
}

func payeeQrPath(uid int) string {
	return fmt.Sprintf("runtime/img/agent/payee_qr_%d.jpg", uid)
}

func absPath(relPath string) string {
	rel := strings.TrimPrefix(filepath.FromSlash(relPath), string(filepath.Separator))
	return filepath.Join(global.GVA_CONFIG.Q取运行目录, rel)
}

func normalizeRuntimeImagePath(raw string) string {
	path := strings.TrimSpace(raw)
	path = strings.TrimPrefix(path, "/")
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	if path == "" || strings.Contains(path, "..") || filepath.IsAbs(path) {
		return ""
	}
	if !strings.HasPrefix(path, "runtime/img/") {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return ""
	}
	return path
}

func imageInfo(relPath string) (WithdrawImageInfo, error) {
	abs := absPath(relPath)
	root := filepath.Clean(filepath.Join(global.GVA_CONFIG.Q取运行目录, "runtime", "img"))
	cleanAbs := filepath.Clean(abs)
	if cleanAbs != root && !strings.HasPrefix(cleanAbs, root+string(filepath.Separator)) {
		return WithdrawImageInfo{}, errors.New("图片地址错误")
	}
	if _, ok := fileExists(cleanAbs); !ok {
		return WithdrawImageInfo{}, errors.New("图片不存在")
	}
	return WithdrawImageInfo{AbsPath: cleanAbs, Ext: strings.ToLower(filepath.Ext(cleanAbs))}, nil
}

func fileExists(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	return info, err == nil && !info.IsDir()
}

func copyFile(src string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func finishTx(tx *gorm.DB, err error) error {
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func normalizeListRequest(req *WithdrawListRequest) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.Size > 100 {
		req.Size = 100
	}
}

func makeOrderNo(uid int) string {
	return fmt.Sprintf("WD%d%d", time.Now().UnixNano(), uid)
}

func randomToken() string {
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%x-%d", buf, time.Now().UnixNano())))
	return hex.EncodeToString(hash[:])
}
