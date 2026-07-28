package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
)

func Ecc_生成密钥对() (私钥Bsse64, 公钥Bsse64 string, err error) {

	// 1. 生成 ECC 密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}

	// 2. 导出私钥为 Base64
	私钥Bsse64, err = exportPrivateKeyToBase64(privateKey)
	if err != nil {
		return
	}
	//fmt.Println("私钥 (Base64):", privateKeyBase64)

	// 3. 导出公钥为 Base64
	公钥Bsse64, err = exportPublicKeyToBase64(&privateKey.PublicKey)
	if err != nil {
		return
	}
	return
}
func Ecc_公钥加密(公钥Base64 string, 明文 []byte) (密文 []byte, err error) {

	var importedPublicKey *ecdsa.PublicKey
	importedPublicKey, err = importPublicKeyFromBase64(公钥Base64) // 5. 从 Base64 导入公钥
	if err != nil {
		return
	}
	// 6. 使用公钥加密数据
	密文, err = encryptWithPublicKey(importedPublicKey, 明文)
	return
}

func Ecc_私钥解密(私钥Base64 string, 密文 []byte) (明文 []byte, err error) {

	var importedPrivateKey *ecdsa.PrivateKey
	// 4. 从 Base64 导入私钥
	importedPrivateKey, err = importPrivateKeyFromBase64(私钥Base64)
	if err != nil {
		return
	}
	// 6. 使用公钥加密数据
	明文, err = decryptWithPrivateKey(importedPrivateKey, 密文)
	return
}

// Ecc_私钥签名 使用私钥对数据进行签名
func Ecc_私钥签名(私钥Base64 string, 数据 []byte) (签名 []byte, err error) {
	// 从 Base64 导入私钥
	importedPrivateKey, err := importPrivateKeyFromBase64(私钥Base64)
	if err != nil {
		return nil, err
	}

	// 计算数据的哈希值
	hash := sha256.Sum256(数据)

	// 使用私钥对哈希值进行签名
	签名, err = ecdsa.SignASN1(rand.Reader, importedPrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	return 签名, nil
}

// Ecc_公钥验签 使用公钥验证签名
func Ecc_公钥验签(公钥Base64 string, 数据 []byte, 签名 []byte) (验证通过 bool, err error) {
	// 从 Base64 导入公钥
	importedPublicKey, err := importPublicKeyFromBase64(公钥Base64)
	if err != nil {
		return false, err
	}

	// 计算数据的哈希值
	hash := sha256.Sum256(数据)

	// 使用公钥验证签名
	验证通过 = ecdsa.VerifyASN1(importedPublicKey, hash[:], 签名)
	return 验证通过, nil
}

// 导出私钥为 Base64
func exportPrivateKeyToBase64(privateKey *ecdsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(privateKeyBytes), nil
}

// 导出公钥为 Base64
func exportPublicKeyToBase64(publicKey *ecdsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(publicKeyBytes), nil
}

// 从 Base64 导入私钥
func importPrivateKeyFromBase64(privateKeyBase64 string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(privateKeyBytes)
}

// 从 Base64 导入公钥
func importPublicKeyFromBase64(publicKeyBase64 string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	return pub.(*ecdsa.PublicKey), nil
}

// 使用公钥加密数据
func encryptWithPublicKey(publicKey *ecdsa.PublicKey, data []byte) ([]byte, error) {
	// 生成临时密钥对
	tempPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// 使用 ECDH 生成共享密钥
	sharedSecretX, _ := publicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, tempPrivateKey.D.Bytes())
	sharedSecret := sha256.Sum256(sharedSecretX.Bytes())

	// 使用 AES 加密数据
	block, err := aes.NewCipher(sharedSecret[:])
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	// 返回临时公钥和加密数据
	tempPublicKey := tempPrivateKey.PublicKey
	tempPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&tempPublicKey)
	if err != nil {
		return nil, err
	}
	return append(tempPublicKeyBytes, ciphertext...), nil
}

// 使用私钥解密数据
func decryptWithPrivateKey(privateKey *ecdsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	// 解析临时公钥
	tempPublicKeyBytes := encryptedData[:91] // P-256 公钥长度为 91 字节
	tempPublicKey, err := x509.ParsePKIXPublicKey(tempPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	// 使用 ECDH 生成共享密钥
	sharedSecretX, _ := tempPublicKey.(*ecdsa.PublicKey).Curve.ScalarMult(tempPublicKey.(*ecdsa.PublicKey).X, tempPublicKey.(*ecdsa.PublicKey).Y, privateKey.D.Bytes())
	sharedSecret := sha256.Sum256(sharedSecretX.Bytes())

	// 使用 AES 解密数据
	block, err := aes.NewCipher(sharedSecret[:])
	if err != nil {
		return nil, err
	}
	ciphertext := encryptedData[91:]
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("密文太短")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}
