// controller/user_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"errors"
	"github.com/gin-gonic/gin"
)

// SwaggerUserResponse 用户信息响应（用于 Swagger 文档）
type SwaggerUserResponse struct {
	ID        uint   `json:"id" example:"1"`
	Username  string `json:"username" example:"testuser"`
	Email     string `json:"email" example:"test@example.com"`
	Avatar    string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Bio       string `json:"bio" example:"个人简介"`
	Role      string `json:"role" example:"user"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// SwaggerLoginData 登录响应数据（用于 Swagger 文档）
type SwaggerLoginData struct {
	Token string              `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  SwaggerUserResponse `json:"user"`
}

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

type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register 用户注册
// @Summary      用户注册
// @Description  注册新用户
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      200  {object}  utils.Response{data=SwaggerLoginData} "注册成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      409  {object}  utils.Response "用户已存在"
// @Router       /auth/register [post]
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

	user, token, err := ctrl.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "注册成功", gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录，返回 JWT token
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200  {object}  utils.Response{data=SwaggerLoginData} "登录成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "用户名或密码错误"
// @Router       /auth/login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	user, token, err := ctrl.userService.Login(req.Username, req.Password)
	if err != nil {
		utils.Error(c, err)
		return
	}

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

// RefreshToken 刷新 Token
// @Summary      刷新 Token
// @Description  刷新过期的 JWT token
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Token 信息"
// @Success      200  {object}  utils.Response{data=object{token=string}} "刷新成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "Token 无效或已过期"
// @Router       /auth/refresh [post]
func (ctrl *UserController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	newToken, err := ctrl.userService.RefreshToken(req.Token)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"token": newToken,
	})
}

// GetProfile 获取用户信息
// @Summary      获取当前用户信息
// @Description  获取已登录用户的个人信息
// @Tags         用户信息
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=SwaggerUserResponse} "获取成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /profile [get]
func (ctrl *UserController) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, errors.New("没有从context中找到userID"))
		return
	}

	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"avatar":     user.Avatar,
		"bio":        user.Bio,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}
