package utils

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"server/utils/Qqwry"
	"server/utils/class"
	"strconv"
	"strings"
	"testing"
	"time"
)

const 测试Rsa私钥 = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCVp+zp4GDurEFlDGXu1gpLwgnpJCxQpv9WBhdlYwQcHVTOM4fH
A54nFP6EWNEBPXUqYRnkAcY0GFPpVg94tCTcGYhNM1YE1xdt20wCnHBigp6Ftp+2
LUDec73d5qqsSXYvuRj44c6sINxFlqpGsqwv/5GFKQA2DDRtFvgodMDSwwIDAQAB
AoGARDPKUW+TXVIdX1q+HZFoEcx1Tc3RcFQa625NPURZvCJV4r8zMqxgZ/k2YMRF
Q+ZpPg4QiYuRc25e12kEFgMrBH0qWNebCF8IOnTx3VDC/0J6CeSRiktx0DEC7dTp
hBaXQeyitz3vMAgpgS50hc47x51Uq2YnrPeEp3LP4wLa4DECQQDEMw50W6eLa+14
acRM4i1Qqkv40XT2Xl/6yJPvuCqhRHEa0HSlx46baLz6u9AI42oWEW6yAtUZWKHI
Omr7toKLAkEAw0UzEn+FMdBLRtx/lazF4jzxCvGn0DmXKNCunoFbNrLF5XIQFlyW
xLQBFo42u1MkNFNdZxSqfTHOyshaWyuPqQJAUc2r4C363HHCkGsg8qF3HWPzyymh
57sPr6ODsvaJp7W+ptx1Vk2vEvMHoE9AfQQ8zji0e7ocmbXPtolM4+iD4wJACnkc
qyJRx4EveGZ9JU718hNorjiV0pI0LEk9BF1VwhJGkO7UNA7VO0mYQVhxUQy9hVzv
RocSOdLBD5k9V2R3uQJAfxKUFRuGi4JmSPDhU8BqmoJWK4tuYypSi8uVlx0b6Suz
EGfbpm/BNVHVGtS5Afay//CDB03B2MbbsKerKPlhoQ==
-----END RSA PRIVATE KEY-----
`
const 测试Rsa公钥 = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCVp+zp4GDurEFlDGXu1gpLwgnp
JCxQpv9WBhdlYwQcHVTOM4fHA54nFP6EWNEBPXUqYRnkAcY0GFPpVg94tCTcGYhN
M1YE1xdt20wCnHBigp6Ftp+2LUDec73d5qqsSXYvuRj44c6sINxFlqpGsqwv/5GF
KQA2DDRtFvgodMDSwwIDAQAB
-----END PUBLIC KEY-----
`

func Test_启动子程序(t *testing.T) {
	测试队列()

	fmt.Println("执行完毕")
}

func Ip查地址测试() {
	ip地址 := "1.57.10.118" // 要查询的IP地址
	局_耗时 := time.Now().UnixMilli()
	for i := 0; i < 10000; i++ {
		_, _ = Qqwry.C查询IP归属地(ip地址)
	}
	fmt.Printf("1w次ip查询耗时:%dms\n", time.Now().UnixMilli()-局_耗时)
}

func Aes加解密测试() {

	局_耗时 := time.Now().UnixNano() / 1e6
	for I := 0; I < 100000; I++ {
		_ = Aes加密_cbc192("123456", "0NQ158xWiLJeIUCh8zhJI7ekpAZPvWlC")
	}
	局_耗时 = time.Now().UnixNano()/1e6 - 局_耗时
	fmt.Println("Aes加密耗时:" + strconv.FormatInt(局_耗时, 10))

	密文 := Aes加密_cbc192("123465", "0NQ158xWiLJeIUCh8zhJI7ekpAZPvWlC")
	fmt.Println(hex.EncodeToString(密文))
	明文 := Aes解密_cbc192(密文, "0NQ158xWiLJeIUCh8zhJI7ekpAZPvWlC")
	fmt.Println(明文)
	fmt.Println("签名_Aes")
	fmt.Println(签名_Aes("OH+62zyTtaDXdnBjzwSmCYBG4JWsq1BtxqjQANVIarRLixwqJ9JPQNWErDfet5WtUGztsr484k3rfCkxBZ6O2gli6SYCizahqxmOJqVRqMcDqJeHlusMA/cdO5j7hLQH8goiOXs6pD1s0HxZ73rGtAYPdqYB7P3nfaZMTRtcbq8=", "0NQ158xWiLJeIUCh8zhJI7ekpAZPvWlC"))
	fmt.Println("签名_Rsa")
	fmt.Println(签名_Rsa("OH+62zyTtaDXdnBjzwSmCYBG4JWsq1BtxqjQANVIarRLixwqJ9JPQNWErDfet5WtUGztsr484k3rfCkxBZ6O2gli6SYCizahqxmOJqVRqMcDqJeHlusMA/cdO5j7hLQH8goiOXs6pD1s0HxZ73rGtAYPdqYB7P3nfaZMTRtcbq8=", 测试Rsa私钥))
}

func Rsa加解密测试() {
	局_明文 := "测试专用明文"
	base64密文 := Rsa公钥加密(测试Rsa公钥, []byte(局_明文))
	局_密文字节集, _ := base64.StdEncoding.DecodeString(base64密文)
	局_明文 = Rsa私钥解密([]byte(测试Rsa私钥), 局_密文字节集)

	局_明文 = "测试专用明文"
	base64密文 = Rsa私钥签名(局_明文, 测试Rsa私钥)

	局_明文 = "测试专用明文"
	base64密文 = RSA私钥加密([]byte(测试Rsa私钥), []byte(局_明文))
	局_密文字节集, _ = base64.StdEncoding.DecodeString(base64密文)
	局_明文 = string(RSA公钥解密(测试Rsa公钥, 局_密文字节集))

}

func Rsa测试() {
	错误, 公钥base64, 私钥base64 := GetRsaKey()

	fmt.Println("私钥:")
	fmt.Println(私钥base64)
	fmt.Println("公钥:")
	fmt.Println(公钥base64)
	fmt.Println(错误)
}

func 签名_Aes(base64后明文 string, AesKey string) string {
	return strings.ToUpper(Md5String(base64后明文 + AesKey))
}

func 签名_Rsa(base64后明文 string, RSA私钥 string) string {

	pemKey := []byte(RSA私钥)

	data := []byte(base64后明文)
	hashMd5 := md5.Sum(data)
	hashed := hashMd5[:]
	block, _ := pem.Decode(pemKey)
	if block == nil {
		return ""
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes) //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,
	if err != nil {
		return ""
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.MD5, hashed)
	return strings.ToUpper(hex.EncodeToString(signature))
}
func 测试队列() {
	var 队列 = class.L_队列{}
	for i := 0; i < 1000; i++ {
		队列.J加入队列(strconv.Itoa(i))
	}
	局_临时文本 := ""
	局_临时文本2 := ""
	for 队列.Q取队列长度() > 0 {
		go func() {
			if 队列.T弹出队列文本(&局_临时文本) {
				局_临时文本2 += (局_临时文本 + "\r\n")
				//fmt.Println("弹出队列:" + 局_临时文本)
			} else {
				//fmt.Println("弹出队列失败")
			}
		}()

	}
	C程序_延时(5000)
	fmt.Println("最终" + strconv.Itoa(W文本_取行数(局_临时文本2)))
	fmt.Println(局_临时文本2)

}
