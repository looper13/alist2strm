package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	letterBytes  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes  = "0123456789"
	specialBytes = "!@#$%^&*"
	passwordLen  = 12
)

// GenerateRandomPassword 生成一个随机密码
func GenerateRandomPassword() string {
	password := make([]byte, passwordLen)

	// 确保至少包含一个小写字母、一个大写字母、一个数字和一个特殊字符
	password[0] = letterBytes[randInt(0, 25)]                   // 小写字母
	password[1] = letterBytes[randInt(26, 51)]                  // 大写字母
	password[2] = numberBytes[randInt(0, len(numberBytes)-1)]   // 数字
	password[3] = specialBytes[randInt(0, len(specialBytes)-1)] // 特殊字符

	// 填充剩余字符
	for i := 4; i < passwordLen; i++ {
		chars := letterBytes + numberBytes + specialBytes
		password[i] = chars[randInt(0, len(chars)-1)]
	}

	// 打乱密码顺序
	for i := len(password) - 1; i > 0; i-- {
		j := randInt(0, i)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

func randInt(min, max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		panic(err)
	}
	return int(n.Int64()) + min
}
