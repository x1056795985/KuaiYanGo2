package main

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
	"fmt"
	"io"
	"log"
)

func TestEcc() {
	// 1. 生成 ECC 密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}

	// 2. 导出私钥为 Base64
	privateKeyBase64, err := exportPrivateKeyToBase64(privateKey)
	if err != nil {
		log.Fatalf("导出私钥失败: %v", err)
	}
	fmt.Println("私钥 (Base64):", privateKeyBase64)

	// 3. 导出公钥为 Base64
	publicKeyBase64, err := exportPublicKeyToBase64(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("导出公钥失败: %v", err)
	}
	fmt.Println("公钥 (Base64):", publicKeyBase64)

	// 4. 从 Base64 导入私钥
	importedPrivateKey, err := importPrivateKeyFromBase64(privateKeyBase64)
	if err != nil {
		log.Fatalf("导入私钥失败: %v", err)
	}

	// 5. 从 Base64 导入公钥
	importedPublicKey, err := importPublicKeyFromBase64(publicKeyBase64)
	if err != nil {
		log.Fatalf("导入公钥失败: %v", err)
	}

	// 6. 使用公钥加密数据
	message := "Hello, ECC!"
	encryptedData, err := encryptWithPublicKey(importedPublicKey, []byte(message))
	if err != nil {
		log.Fatalf("加密失败: %v", err)
	}
	fmt.Println("加密后的数据 (Base64):", base64.StdEncoding.EncodeToString(encryptedData))

	// 7. 使用私钥解密数据
	decryptedData, err := decryptWithPrivateKey(importedPrivateKey, encryptedData)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
	}
	fmt.Println("解密后的数据:", string(decryptedData))
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
