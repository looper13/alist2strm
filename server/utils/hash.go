package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
)

// MD5 计算字符串的MD5哈希值
func MD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA256 计算字符串的SHA256哈希值
func SHA256(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD5Reader 计算io.Reader的MD5哈希值
func MD5Reader(r io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// SHA256Reader 计算io.Reader的SHA256哈希值
func SHA256Reader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
