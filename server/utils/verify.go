package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	if length < 8 {
		length = 8
	}

	password := make([]byte, length)
	for i := range password {
		// 使用加密安全的随机数生成器
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果加密随机数生成失败，回退到简单方式
			password[i] = charset[i%len(charset)]
		} else {
			password[i] = charset[num.Int64()]
		}
	}

	return string(password)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// IsValidStatus 检查状态是否有效
func IsValidStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "pending", "suspended"}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

// SanitizeString 清理字符串
func SanitizeString(str string) string {
	// 移除前后空格
	str = strings.TrimSpace(str)

	// 移除多余的空格
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")

	return str
}
