// controller/post_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService *service.PostService
}

func NewPostController(postService *service.PostService) *PostController {
	return &PostController{postService: postService}
}

type CreatePostRequest struct {
	Title      string   `json:"title" binding:"required,min=1,max=200"`
	Content    string   `json:"content" binding:"required"`
	CategoryID uint     `json:"category_id"`
	Tags       []string `json:"tags"`
}

type UpdatePostRequest struct {
	Title      string `json:"title" binding:"required,min=1,max=200"`
	Content    string `json:"content" binding:"required"`
	CategoryID uint   `json:"category_id"`
}

func (ctrl *PostController) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	// 从 JWT 中获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, errors.New("从JWT中获取当前用户ID错误"))
		return
	}

	post, err := ctrl.postService.Create(req.Title, req.Content, userID, req.CategoryID, req.Tags)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章创建成功", post)
}

func (ctrl *PostController) GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	post, err := ctrl.postService.GetPostByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, post)
}

// UpdatePost 更新文章（需要检查权限）
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	// 获取当前用户信息
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	role, _ := middleware.GetUserRole(c)

	// 获取文章信息，检查权限
	post, err := ctrl.postService.GetPostByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	// 权限检查：只有作者本人或管理员可以更新
	if post.UserID != userID && role != "admin" {
		utils.Error(c, utils.ErrInsufficientPermission)
		return
	}

	err = ctrl.postService.UpdatePost(uint(id), req.Title, req.Content, req.CategoryID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章更新成功", nil)
}

// DeletePost 删除文章（需要检查权限）
func (ctrl *PostController) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	// 获取当前用户信息
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	role, _ := middleware.GetUserRole(c)

	// 获取文章信息，检查权限
	post, err := ctrl.postService.GetPostByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	// 权限检查：只有作者本人或管理员可以删除
	if post.UserID != userID && role != "admin" {
		utils.Error(c, utils.ErrInsufficientPermission)
		return
	}

	err = ctrl.postService.DeletePost(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章删除成功", nil)
}

// ListPosts 获取文章列表（公开，但管理员可以看到所有状态）
func (ctrl *PostController) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "1"))
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 32)

	// 获取用户角色（可选认证）
	role, _ := middleware.GetUserRole(c)

	// 如果不是管理员，只显示已发布的文章
	if role != "admin" {
		status = 1 // 强制只显示已发布
	}

	posts, total, err := ctrl.postService.ListPosts(page, pageSize, status, uint(categoryID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        posts,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
