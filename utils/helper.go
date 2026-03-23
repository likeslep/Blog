// utils/helper.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

// 生成随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:length]
}

// 验证邮箱格式
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// 验证用户名格式
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	pattern := `^[a-zA-Z0-9_]+$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// 清理HTML标签（简单实现）
func StripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// 截取字符串
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
