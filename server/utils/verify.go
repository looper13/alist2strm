package utils

import (
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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length < 8 {
		length = 8
	}

	password := make([]byte, length)
	for i := range password {
		password[i] = charset[i%len(charset)]
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
