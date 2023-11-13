package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// 加密
func Aes加密_cbc192(orig string, key string) []byte {
	// 转成字节数组
	if len(key) != 24 {
		return []byte{}
	}
	origData := []byte(orig)
	k := []byte(key)[:24]
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补 全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式  iv
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return cryted
}

// 加密
func Aes加密_cbc192_2(加密数据 []byte, key string) []byte {
	// 转成字节数组
	if len(key) != 24 {
		return []byte{}
	}
	origData := 加密数据
	k := []byte(key)[:24]
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补 全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式  iv
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return cryted
}
func Aes加密_cbc192密匙字节数组(orig string, key []byte) []byte {
	// 转成字节数组
	if len(key) != 24 {
		return []byte{}
	}
	origData := []byte(orig)
	k := key
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补 全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式  iv
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return cryted
}
func Aes解密_cbc192(加密数据 []byte, key string) string {
	return Aes解密_cbc192字节集(加密数据, []byte(key))
}
func Aes解密_cbc192字节集(加密数据 []byte, key []byte) string {
	if len(key) != 24 || len(加密数据) < 16 {
		return ""
	}
	k := key[:24]
	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	//blockSize := block.BlockSize()
	// 加密模式
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// 创建数组
	orig := make([]byte, len(加密数据))
	// 解密
	blockMode.CryptBlocks(orig, 加密数据)
	// 去补 全码
	orig = PKCS7UnPadding(orig)
	return string(orig)
}
func Aes解密_cbc192字节集2(加密数据 []byte, key []byte) []byte {
	if len(key) != 24 || len(加密数据) < 16 {
		return []byte{}
	}
	k := key[:24]
	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	//blockSize := block.BlockSize()
	// 加密模式
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// 创建数组
	orig := make([]byte, len(加密数据))
	// 解密
	blockMode.CryptBlocks(orig, 加密数据)
	// 去补 全码
	orig = PKCS7UnPadding(orig)
	return orig
}

// 补码
// AES加密数据块分组长度必须为128bit(byte[16])，密钥长度可以是128bit(byte[16])、192bit(byte[24])、256bit(byte[32])中的任意一个。
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 去码
func PKCS7UnPadding(origData []byte) []byte {

	length := len(origData)
	if length-1 < 0 {
		return origData
	}
	unpadding := int(origData[length-1])
	if length-unpadding < 0 {
		fmt.Printf("AESPKCS7UnPadding 数据异常:%v", origData)
		fmt.Printf("AESPKCS7UnPadding 数据异常:%s", origData)
		return origData
	}
	return origData[:(length - unpadding)]
}
