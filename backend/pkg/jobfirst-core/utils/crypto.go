package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SHA256Hash 使用SHA256哈希
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateToken 生成随机token
func GenerateToken(length int) (string, error) {
	return GenerateRandomString(length)
}

// Encrypt 简单的加密函数（实际项目中应使用更安全的加密方法）
func Encrypt(data, key string) (string, error) {
	// 这里使用简单的XOR加密，实际项目中应使用AES等安全加密算法
	encrypted := make([]byte, len(data))
	keyBytes := []byte(key)

	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ keyBytes[i%len(keyBytes)]
	}

	return hex.EncodeToString(encrypted), nil
}

// Decrypt 简单的解密函数
func Decrypt(encryptedData, key string) (string, error) {
	data, err := hex.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(data))
	keyBytes := []byte(key)

	for i := 0; i < len(data); i++ {
		decrypted[i] = data[i] ^ keyBytes[i%len(keyBytes)]
	}

	return string(decrypted), nil
}

// GenerateUUID 生成UUID（简化版本）
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
