package captcha

import (
	"encoding/base64"
	"errors"
	"image"
	"math/rand"
	"strconv"
	"strings"

	"server/app/logic/common/captcha/clickItem"
)

// 集_点击题型表 已注册的题型成员及其抽取权重：
// 权重即灰度机制，新题型以小权重上线观察，被攻破的题型可动态降权。
var 集_点击题型表 = []结构_题型项{
	{题型: clickItem.T题_静态扰动{}, 权重: 6},
	{题型: clickItem.T题_遮挡脑补{}, 权重: 3},
}

// 结构_题型项 注册表条目：题型实现与抽取权重。
type 结构_题型项 struct {
	题型 点击_题型
	权重 int
}

// 点击_题型 点击式验证码题型成员的统一契约。
// 所有题型的产物都归一化为"画布帧序列+有序目标矩形"，
// 因此验证协议、存储格式与前端交互完全不变。
// 新增题型只需在 clickItem 包实现本接口并注册到 集_点击题型表。
type 点击_题型 interface {
	// Q题型_名称 返回题型名称，用于日志与失败率统计。
	Q题型_名称() string
	// T题_生成 按难度生成一题，随机源由入口层注入(crypto种子)。
	T题_生成(难度 int, 随机 *rand.Rand) (*clickItem.T题_挑战, error)
}

//	创建一个按序点击挑战：从注册成员中加权随机抽取题型生成，
//
// 攻击者必须同时攻破所有已注册题型。difficulty 控制干扰图标数量。
func GenerateClick(difficulty int) (id, dataURL string, err error) {
	difficulty = 难度_钳制(difficulty)

	id, err = randomToken(18, "23456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ")
	if err != nil {
		return "", "", err
	}
	局_随机, err := clickItem.X渲染_新建随机源()
	if err != nil {
		return "", "", err
	}
	局_挑战, err := 题型_抽取(difficulty, 局_随机)
	if err != nil {
		return "", "", err
	}

	局_前缀, 局_数据, err := 编码_挑战图(局_挑战)
	if err != nil {
		return "", "", err
	}
	VerificationCodes.set(id, 局_挑战.M目标)
	return id, 局_前缀 + base64.StdEncoding.EncodeToString(局_数据), nil
}

// VerifyClick 校验四个有序的 x|y 坐标点，consume 为真时消费该挑战。
func VerifyClick(id, answer string, consume bool) bool {
	if id == "" || answer == "" {
		return false
	}
	unlock := VerificationCodes.lock(id)
	defer unlock()

	raw, ok := VerificationCodes.getLocked(id)
	if !ok {
		return false
	}
	targets, ok := raw.([]image.Rectangle)
	if !ok || len(targets) != clickItem.T题_点击数 {
		return false
	}
	points := strings.Split(answer, ",")
	if len(points) != clickItem.T题_点击数 {
		return false
	}
	for i, encoded := range points {
		xText, yText, ok := strings.Cut(encoded, "|")
		if !ok {
			return false
		}
		x, xErr := strconv.Atoi(xText)
		y, yErr := strconv.Atoi(yText)
		if xErr != nil || yErr != nil || !image.Pt(x, y).In(targets[i]) {
			return false
		}
	}
	if consume {
		VerificationCodes.deleteLocked(id)
	}
	return true
}

// 难度_钳制 将难度限制到题型支持的区间。
func 难度_钳制(难度 int) int {
	if 难度 < 0 {
		return 0
	}
	if 难度 > clickItem.T题_最大难度 {
		return clickItem.T题_最大难度
	}
	return 难度
}

// 题型_抽取 按权重随机抽一个成员生成挑战；
// 成员生成失败或产物非法时降级到静态成员，保证可用性。
func 题型_抽取(难度 int, 随机 *rand.Rand) (*clickItem.T题_挑战, error) {
	局_总权重 := 0
	for _, 局_项 := range 集_点击题型表 {
		局_总权重 += 局_项.权重
	}
	局_抽中 := 随机.Intn(局_总权重)
	局_题型 := 集_点击题型表[len(集_点击题型表)-1].题型
	for _, 局_项 := range 集_点击题型表 {
		if 局_抽中 < 局_项.权重 {
			局_题型 = 局_项.题型
			break
		}
		局_抽中 -= 局_项.权重
	}

	局_挑战, 局_错误 := 局_题型.T题_生成(难度, 随机)
	if 局_错误 == nil && 题_合法(局_挑战) {
		return 局_挑战, nil
	}
	if 局_挑战, 局_错误 = (clickItem.T题_静态扰动{}).T题_生成(难度, 随机); 局_错误 != nil {
		return nil, errors.New("点击验证码题型生成失败: " + 局_错误.Error())
	}
	return 局_挑战, nil
}

// 题_合法 校验挑战产物的基本完整性。
func 题_合法(挑战 *clickItem.T题_挑战) bool {
	return 挑战 != nil && len(挑战.Z帧) > 0 && len(挑战.M目标) == clickItem.T题_点击数
}
