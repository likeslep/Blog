// controller/post_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
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

	// TODO: 从JWT中获取当前用户ID
	userID := uint(1) // 临时写死

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

	err = ctrl.postService.UpdatePost(uint(id), req.Title, req.Content, req.CategoryID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章更新成功", nil)
}

func (ctrl *PostController) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	err = ctrl.postService.DeletePost(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章删除成功", nil)
}

func (ctrl *PostController) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "1"))
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 32)

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
