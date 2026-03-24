package middleware

import (
	"blog/utils"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// 认证头
	AuthorizationHeader = "Authorization"
	// Token 前缀
	TokenPrefix = "Bearer"
	// 上下文中的用户信息键
	UserIDKey   = "user_id"
	UsernameKey = "username"
	UserRoleKey = "user_role"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization 头
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			utils.Error(c, utils.ErrTokenNotFound)
			c.Abort()
			return
		}

		// 清理可能的换行符和空格
		authHeader = strings.TrimSpace(authHeader)

		// 检查 Token 格式
		if !strings.HasPrefix(authHeader, TokenPrefix) {
			utils.Error(c, errors.New("token 格式不正确"))
			c.Abort()
			return
		}

		// 提取 Token
		tokenString := strings.TrimPrefix(authHeader, TokenPrefix)
		tokenString = strings.TrimSpace(tokenString) // 去除可能的空格和换行
		if tokenString == "" {
			utils.Error(c, errors.New("获取token失败"))
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := utils.ParseToken(tokenString, secret)
		if err != nil {
			utils.Error(c, err)
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// 可选中间件：如果提供了 Token 则解析，否则继续
func OptionalJWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader != "" && strings.HasPrefix(authHeader, TokenPrefix) {
			tokenString := strings.TrimPrefix(authHeader, TokenPrefix)
			if claims, err := utils.ParseToken(tokenString, secret); err == nil {
				c.Set(UserIDKey, claims.UserID)
				c.Set(UsernameKey, claims.Username)
				c.Set(UserRoleKey, claims.Role)
			}
		}
		c.Next()
	}
}

// 权限检查中间件
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleKey)
		if !exists {
			utils.Error(c, utils.ErrInsufficientPermission)
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		utils.Error(c, utils.ErrInsufficientPermission)
		c.Abort()
	}
}

// 辅助函数：从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// 辅助函数：从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}
	return username.(string), true
}

// 辅助函数：从上下文获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}
