// utils/response.go
package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 统一响应结构
type Response struct {
	Code    int         `json:"code" example:"0"`          // 状态码，0表示成功
	Message string      `json:"message" example:"success"` // 响应消息
	Data    interface{} `json:"data,omitempty"`            // 响应数据
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// 错误响应
func Error(c *gin.Context, err error) {
	httpStatus := GetHTTPStatus(err)
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: err.Error(),
	})
}

// 自定义错误响应
func ErrorWithMessage(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
	})
}

// 参数验证错误响应
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}
