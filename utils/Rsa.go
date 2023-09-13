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
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func Rsa私钥签名(base64后明文 string, RSA私钥 string) string {

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
	// 感觉和私钥加密区别实际就是 hash 参数有区别, 如果没这个参数就是私钥加密明文,如果有这个参数就是私钥加密 hash(明文)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.MD5, hashed)
	return strings.ToUpper(hex.EncodeToString(signature))
}

// RSA公钥私钥产生
func GetRsaKey() (err error, PublicKey string, PrivateKey string) {
	bits := 1024
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err, "", ""
	}

	privateKeyStream := x509.MarshalPKCS1PrivateKey(privateKey)
	//file, err := os.Create("./key/private.pem")
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyStream,
	}
	//_ = pem.Encode(file, block)
	privateKeyStr := string(pem.EncodeToMemory(block))
	publicKey := &privateKey.PublicKey
	publicKeyStream, _ := x509.MarshalPKIXPublicKey(publicKey)
	//publicKeyStream := x509.MarshalPKCS1PublicKey(publicKey)
	//file, err = os.Create("./key/public.pem")
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyStream,
	}
	//_ = pem.Encode(file, block)
	publicKeyStr := string(pem.EncodeToMemory(block))
	return err, publicKeyStr, privateKeyStr
}

func Rsa私钥解密(Rsa私钥 []byte, 加密数据 []byte) string {
	return string(Rsa私钥解密2(Rsa私钥, 加密数据))
}
func Rsa私钥解密2(Rsa私钥 []byte, 加密数据 []byte) []byte {

	if len(加密数据) == 0 || len(Rsa私钥) == 0 {
		return []byte{}
	}

	//2、pem decode,得到block的der编码数据
	block, _ := pem.Decode(Rsa私钥)
	if block == nil {
		fmt.Printf("私钥载入失败可能格式不正确:%s\n", string(Rsa私钥))
		return []byte{}
		//密钥错误
	}
	derText := block.Bytes //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,
	//3、解码der得到私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(derText)
	if err != nil {
		return []byte{}
		//密钥错误
	}
	//4、私钥解密
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, 加密数据)

	if err != nil {
		return []byte{}
		//解密失败
	}
	return plainText

}
func Rsa公钥加密(公钥 string, 加密内容 []byte) (base64密文 string) {
	//2、pem decode,得到block的der编码数据
	block, _ := pem.Decode([]byte(公钥))
	if block == nil {
		fmt.Printf("Rsa公钥加密公钥载入失败可能格式不正确:%s\n", 公钥)
		return ""
		//密钥错误
	}
	derText := block.Bytes //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,
	//3、解码der得到公钥
	//publicKey, err := x509.ParsePKCS1PublicKey(derText)   // 这个只能载入PKCS1的  但是目前公钥基本都是PKCS8的

	publicKeyInterface, err := x509.ParsePKIXPublicKey(derText) //这个只能载入PKCS8的公钥
	if err != nil {
		return ""
		//密钥错误
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey) //强制转换

	//4、公钥加密
	cipherData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, 加密内容)
	if err != nil {
		return
	}
	base64密文 = base64.StdEncoding.EncodeToString(cipherData)
	return base64密文

}

// 参考地址 https://www.cnblogs.com/imlgc/p/7076313.html
func RSA私钥加密(Rsa私钥 []byte, 明文 []byte) (base64密文 string) {
	if len(明文) == 0 || len(Rsa私钥) == 0 {
		return ""
	}

	//2、pem decode,得到block的der编码数据

	block, _ := pem.Decode(Rsa私钥)
	if block == nil {
		fmt.Printf("RSA私钥加密Rsa私钥载入失败可能格式不正确:%s\n", string(Rsa私钥))
		return ""
		//密钥错误
	}
	derText := block.Bytes //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,
	//3、解码der得到私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(derText)

	signData, err := rsa.SignPKCS1v15(nil, privateKey, crypto.Hash(0), 明文)
	if err != nil {
		return ""
	}
	base64密文 = base64.StdEncoding.EncodeToString(signData)
	return base64密文
}
func RSA公钥解密(公钥 string, 密文 []byte) (明文字节集 []byte) {

	//2、pem decode,得到block的der编码数据
	block, _ := pem.Decode([]byte(公钥))
	if block == nil {
		fmt.Printf("RSA公钥解密Rsa公钥载入失败可能格式不正确:%s\n", 公钥)
		return []byte{}
		//密钥错误
	}
	derText := block.Bytes //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,

	//3、解码der得到公钥

	publicKey, err := x509.ParsePKCS1PublicKey(derText)
	//3、解码der得到公钥
	//publicKey, err := x509.ParsePKCS1PublicKey(derText)   // 这个只能载入PKCS1的  但是目前公钥基本都是PKCS8的

	publicKeyInterface, err := x509.ParsePKIXPublicKey(derText) //这个只能载入PKCS8的公钥
	if err != nil {
		return 明文字节集
		//密钥错误
	}

	publicKey = publicKeyInterface.(*rsa.PublicKey) //强制转换

	明文字节集, err = publicDecrypt(publicKey, crypto.Hash(0), nil, 密文)
	if err != nil {
		return 明文字节集
	}
	return 明文字节集
}

// copy from crypt/rsa/pkcs1v5.go
var hashPrefixes = map[crypto.Hash][]byte{
	crypto.MD5:       {0x30, 0x20, 0x30, 0x0c, 0x06, 0x08, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x02, 0x05, 0x05, 0x00, 0x04, 0x10},
	crypto.SHA1:      {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA224:    {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
	crypto.SHA256:    {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384:    {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512:    {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
	crypto.MD5SHA1:   {}, // A special TLS case which doesn't use an ASN1 prefix.
	crypto.RIPEMD160: {0x30, 0x20, 0x30, 0x08, 0x06, 0x06, 0x28, 0xcf, 0x06, 0x03, 0x00, 0x31, 0x04, 0x14},
}

// copy from crypt/rsa/pkcs1v5.go
func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}

// copy from crypt/rsa/pkcs1v5.go
func pkcs1v15HashInfo(hash crypto.Hash, inLen int) (hashLen int, prefix []byte, err error) {
	// Special case: crypto.Hash(0) is used to indicate that the data is
	// signed directly.
	if hash == 0 {
		return inLen, nil, nil
	}

	hashLen = hash.Size()
	if inLen != hashLen {
		return 0, nil, errors.New("crypto/rsa: input must be hashed message")
	}
	prefix, ok := hashPrefixes[hash]
	if !ok {
		return 0, nil, errors.New("crypto/rsa: unsupported hash function")
	}
	return
}

// copy from crypt/rsa/pkcs1v5.go
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}
func unLeftPad(input []byte) (out []byte) {
	n := len(input)
	t := 2
	for i := 2; i < n; i++ {
		if input[i] == 0xff {
			t = t + 1
		} else {
			if input[i] == input[0] {
				t = t + int(input[1])
			}
			break
		}
	}
	out = make([]byte, n-t)
	copy(out, input[t:])
	return
}

// copy&modified from crypt/rsa/pkcs1v5.go
func publicDecrypt(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) (out []byte, err error) {
	hashLen, prefix, err := pkcs1v15HashInfo(hash, len(hashed))
	if err != nil {
		return nil, err
	}

	tLen := len(prefix) + hashLen
	k := (pub.N.BitLen() + 7) / 8
	if k < tLen+11 {
		return nil, fmt.Errorf("length illegal")
	}

	c := new(big.Int).SetBytes(sig)
	m := encrypt(new(big.Int), pub, c)
	em := leftPad(m.Bytes(), k)
	out = unLeftPad(em)

	err = nil
	return
}
