package utils

import (
	"regexp"
	"strings"
)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) bool {
	// 用户名：3-20位，只能包含字母、数字、下划线
	pattern := `^[a-zA-Z0-9_]{3,20}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) bool {
	// 密码：至少8位，包含字母和数字
	if len(password) < 8 {
		return false
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasLetter && hasNumber
}

// ValidateCron 验证Cron表达式格式
func ValidateCron(cron string) bool {
	if cron == "" {
		return true // 允许空值
	}

	// 简单的cron表达式验证（5或6个字段）
	fields := strings.Fields(cron)
	return len(fields) == 5 || len(fields) == 6
}

// ValidatePath 验证路径格式
func ValidatePath(path string) bool {
	if path == "" {
		return false
	}

	// 路径不能包含非法字符
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	return true
}

// ValidateFileSuffix 验证文件后缀格式
func ValidateFileSuffix(suffix string) bool {
	if suffix == "" {
		return false
	}

	// 文件后缀应该以逗号分隔，每个后缀以点开头
	suffixes := strings.Split(suffix, ",")
	for _, s := range suffixes {
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, ".") || len(s) < 2 {
			return false
		}
	}

	return true
}
