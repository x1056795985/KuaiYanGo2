package VMP

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"math/big"
	"math/rand"
	"server/new/app/models/common"
	"time"
)

var L_VMP VMP

func init() {
	L_VMP = VMP{}

}

type VMP struct {
}

// nBits Rsa位数 1024
// vPrivate RsaBase64私钥
// vModulus RsaBase64模数
// vProductCode base64产品代码
// /耗时：7 ms  实测速度极快,万次才7毫秒 测试配置 16H40G
func (j *VMP) J计算授权码(c *gin.Context, VmpRsa common.VmpRsa, 授权信息 common.VmpParams) (授权码 string, err error) {
	exported := struct {
		Algorithm   string
		Bits        int
		PrivateKey  string
		Modulus     string
		ProductCode string
	}{
		Algorithm:   "RSA",
		Bits:        VmpRsa.Rsa位数,
		PrivateKey:  VmpRsa.RsaBase64私钥,
		Modulus:     VmpRsa.RsaBase64模数,
		ProductCode: VmpRsa.Base64产品代码,
	}

	serialData, err2 := 打包数据(授权信息, exported.ProductCode)
	if err2 != nil {
		err = err2
		return
	}
	serialBin := serialData

	hash := sha1.Sum(serialBin)
	serialBin = append(serialBin, 0xFF)
	for i := 3; i >= 0; i-- {
		serialBin = append(serialBin, hash[i])
	}

	// 生成 8 到 12 之间的随机整数
	//$size = rand(8, 16);
	//for ($i = 0; $i < $size; $i++) $padding_front[] = rand(1, 255);
	// 改为固定值（示例用全0填充）
	paddingFront := make([]byte, 8)
	for i := range paddingFront {
		paddingFront[i] = byte(rand.Int31n(255) + 1)
	}

	paddingFront = append([]byte{0, 2}, paddingFront...)
	paddingFront = append(paddingFront, 0)

	contentSize := len(serialBin) + len(paddingFront)
	rest := exported.Bits/8 - contentSize
	if rest < 0 {
		err = errors.New("内容太大")
		return
	}

	paddingBack := make([]byte, rest)
	finalSerial := append(paddingFront, serialBin...)
	finalSerial = append(finalSerial, paddingBack...)

	n, _ := base64.StdEncoding.DecodeString(exported.Modulus)
	d, _ := base64.StdEncoding.DecodeString(exported.PrivateKey)

	encoded := S十进制编码(finalSerial)
	modulus := S十进制编码(n)
	private := S十进制编码(d)

	res := new(big.Int).Exp(encoded, private, modulus) //	 模幂运算
	resBytes := S十进制解码(res)
	授权码 = base64.StdEncoding.EncodeToString(resBytes)

	return
}

// 辅助函数
func S十进制编码(data []byte) *big.Int {
	result := big.NewInt(0)
	for _, b := range data {
		result.Mul(result, big.NewInt(256))
		result.Add(result, big.NewInt(int64(b)))
	}
	return result
}

func S十进制解码(n *big.Int) []byte {
	result := []byte{}
	zero := big.NewInt(0)
	twoFiftySix := big.NewInt(256)

	for n.Cmp(zero) > 0 {
		mod := new(big.Int)
		mod.Mod(n, twoFiftySix)
		result = append([]byte{byte(mod.Int64())}, result...)
		n.Div(n, twoFiftySix)
	}
	return result
}

func 打包数据(params common.VmpParams, productCode string) (结果数据 []byte, err error) {
	serial := []byte{1, 1} // Version
	// 用户名
	if len(params.UserName) > 255 {
		err = errors.New("用户名太长")
		return
	}
	serial = append(serial, 2, byte(len(params.UserName)))
	serial = append(serial, []byte(params.UserName)...)

	// Email
	if len(params.Email) > 255 {
		err = errors.New("E-Mail 太长")
		return
	}
	serial = append(serial, 3, byte(len(params.Email)))
	serial = append(serial, []byte(params.Email)...)

	// Hardware ID
	hwid, _ := base64.StdEncoding.DecodeString(params.Hwid)
	if len(hwid) == 0 || len(hwid) > 255 || len(hwid)%4 != 0 {
		err = errors.New("机器码 无效")
		return
	}
	serial = append(serial, 4, byte(len(hwid)))
	serial = append(serial, hwid...)

	// Expire Date
	serial = append(serial, 5)
	serial = appendDate(serial, params.ExpireDate)

	// Time Limit
	if params.TimeLimit < 0 || params.TimeLimit > 255 {
		err = errors.New("无效的时限")
		return
	}
	serial = append(serial, 6, byte(params.TimeLimit))

	// Product Code
	pc, _ := base64.StdEncoding.DecodeString(productCode)
	if len(pc) != 8 {
		err = errors.New("产品代码无效")
		return
	}
	serial = append(serial, 7)
	serial = append(serial, pc...)

	// User Data
	if len(params.UserData) > 255 {
		err = errors.New("用户数据过长")
		return
	}
	serial = append(serial, 8, byte(len(params.UserData)))
	serial = append(serial, params.UserData...)

	// Max Build Date
	serial = append(serial, 9)
	serial = appendDate(serial, params.MaxBuildDate)
	结果数据 = serial
	return
}

func appendDate(serial []byte, a common.S时间) []byte {
	var t time.Time
	t = time.Date(a.Year, time.Month(a.Month), a.Day, 0, 0, 0, 0, time.UTC)
	year, month, day := t.Date()
	return append(serial, byte(day), byte(month), byte(year%256), byte(year/256))

}
