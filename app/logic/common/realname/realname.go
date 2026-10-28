package realname

import (
	"errors"
	"fmt"
	"server/app/logic/common/setting"
	"server/app/utils"
	"strings"
)

// 实名认证平台 与系统设置-实名接口配置.当前选择对应
const (
	S实名平台_未开启 = 0
	S实名平台_腾讯云 = 1
)

// I实名认证 实名认证平台接口,后续对接其他平台时实现此接口,并在Q实名认证实现中注册即可
type I实名认证 interface {
	// 认证 核验姓名与身份证号是否一致,核验一致时返回平台侧唯一请求标识
	R认证(姓名 string, 身份证号 string) (唯一Id string, err error)
}

// S实名_认证 按系统设置选择的实名认证平台核验姓名与身份证号
func S实名_认证(姓名 string, 身份证号 string) (唯一Id string, err error) {
	局_配置 := setting.Q实名认证配置()
	if 局_配置.D当前选择 == S实名平台_未开启 {
		return "", errors.New("管理员未开启实名认证功能")
	}

	姓名 = strings.TrimSpace(姓名)
	身份证号 = strings.TrimSpace(身份证号)
	局_msg := "姓名格式不正确"
	if !utils.Z正则_校验姓名(姓名, &局_msg) {
		return "", errors.New(局_msg)
	}
	局_msg = "身份证号格式不正确"
	if !utils.Z正则_校验身份证号(身份证号, &局_msg) {
		return "", errors.New(局_msg)
	}

	局_实现, err := Q实名认证实现(局_配置.D当前选择)
	if err != nil {
		return "", err
	}
	return 局_实现.R认证(姓名, 身份证号)
}

// Q实名认证实现 按平台id返回实名认证接口实现
func Q实名认证实现(平台 int) (I实名认证, error) {
	switch 平台 {
	case S实名平台_腾讯云:
		return &腾讯云实名认证{}, nil
	default:
		return nil, fmt.Errorf("未知的实名认证平台:%d", 平台)
	}
}
