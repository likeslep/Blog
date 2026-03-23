// controller/user_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *UserController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	// 额外验证
	if !utils.ValidateUsername(req.Username) {
		utils.ValidationError(c, "用户名只能包含字母、数字和下划线")
		return
	}

	user, err := ctrl.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "注册成功", gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func (ctrl *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	user, err := ctrl.userService.Login(req.Username, req.Password)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// TODO: 生成JWT token
	token := "temp_token" // 临时token

	utils.SuccessWithMessage(c, "登录成功", gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (ctrl *UserController) GetProfile(c *gin.Context) {
	// TODO: 从JWT中获取用户ID
	userID := uint(1) // 临时写死

	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
		"bio":      user.Bio,
		"role":     user.Role,
	})
}
