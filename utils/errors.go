// utils/errors.go
package utils

import (
	"errors"
	"net/http"
)

// 自定义错误类型
var (
	// 通用错误
	ErrInternalServer = errors.New("内部服务器错误")
	ErrBadRequest     = errors.New("请求参数错误")
	ErrNotFound       = errors.New("资源不存在")

	// 用户相关错误
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserAlreadyExists  = errors.New("用户已存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrInvalidPassword    = errors.New("用户名或密码错误")
	ErrUserDisabled       = errors.New("账户已被禁用")

	// 文章相关错误
	ErrPostNotFound = errors.New("文章不存在")
	ErrPostExists   = errors.New("文章已存在")

	// 分类相关错误
	ErrCategoryNotFound = errors.New("分类不存在")
	ErrCategoryExists   = errors.New("分类已存在")
	ErrCategoryHasPosts = errors.New("该分类下还有文章，无法删除")

	// 标签相关错误
	ErrTagNotFound = errors.New("标签不存在")
	ErrTagExists   = errors.New("标签已存在")

	// 评论相关错误
	ErrCommentNotFound = errors.New("评论不存在")
	ErrCommentDisabled = errors.New("评论已被禁用")

	// JWT 相关错误
	ErrTokenNotFound          = errors.New("未提供认证令牌")
	ErrInvalidToken           = errors.New("无效的认证令牌")
	ErrTokenExpired           = errors.New("认证令牌已过期")
	ErrTokenNotExpired        = errors.New("令牌未过期，无需刷新")
	ErrInsufficientPermission = errors.New("权限不足")
)

// 错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// 获取错误对应的HTTP状态码
func GetHTTPStatus(err error) int {
	switch err {
	case ErrUserNotFound, ErrPostNotFound, ErrCategoryNotFound,
		ErrTagNotFound, ErrCommentNotFound:
		return http.StatusNotFound
	case ErrUserAlreadyExists, ErrEmailAlreadyExists, ErrCategoryExists, ErrTagExists:
		return http.StatusConflict
	case ErrInvalidPassword, ErrUserDisabled:
		return http.StatusUnauthorized
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// 包装错误信息（用于调试）
func WrapError(err error, message string) error {
	if message == "" {
		return err
	}
	return errors.New(message + ": " + err.Error())
}
