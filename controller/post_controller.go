// controller/post_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SwaggerPostResponse 文章响应（用于 Swagger 文档）
type SwaggerPostResponse struct {
	ID           uint                     `json:"id" example:"1"`
	Title        string                   `json:"title" example:"Go语言入门"`
	Slug         string                   `json:"slug" example:"go-language-intro"`
	Summary      string                   `json:"summary" example:"文章摘要"`
	Content      string                   `json:"content" example:"# 标题\n\n内容..."`
	CoverImage   string                   `json:"cover_image" example:"https://example.com/cover.jpg"`
	ViewCount    int                      `json:"view_count" example:"100"`
	LikeCount    int                      `json:"like_count" example:"10"`
	CommentCount int                      `json:"comment_count" example:"5"`
	Status       int                      `json:"status" example:"1"`
	IsTop        bool                     `json:"is_top" example:"false"`
	UserID       uint                     `json:"user_id" example:"1"`
	CategoryID   uint                     `json:"category_id" example:"1"`
	User         *SwaggerUserResponse     `json:"user,omitempty"`
	Category     *SwaggerCategoryResponse `json:"category,omitempty"`
	Tags         []SwaggerTagResponse     `json:"tags,omitempty"`
	CreatedAt    string                   `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    string                   `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// SwaggerPostListData 文章列表响应（用于 Swagger 文档）
type SwaggerPostListData struct {
	List       []SwaggerPostResponse `json:"list"`
	Total      int64                 `json:"total" example:"100"`
	Page       int                   `json:"page" example:"1"`
	PageSize   int                   `json:"page_size" example:"10"`
	TotalPages int64                 `json:"total_pages" example:"10"`
}

// SwaggerTagListData 标签列表响应（用于 Swagger 文档）
type SwaggerTagListData struct {
	List       []SwaggerTagResponse `json:"list"`
	Total      int64                `json:"total" example:"100"`
	Page       int                  `json:"page" example:"1"`
	PageSize   int                  `json:"page_size" example:"10"`
	TotalPages int64                `json:"total_pages" example:"10"`
}

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
	Title      string   `json:"title" binding:"required,min=1,max=200"`
	Content    string   `json:"content" binding:"required"`
	CategoryID uint     `json:"category_id"`
	Tags       []string `json:"tags"`
}

// Create 创建文章
// @Summary      创建文章
// @Description  创建新文章（需要认证）
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreatePostRequest true "文章信息"
// @Success      200  {object}  utils.Response{data=SwaggerPostResponse} "创建成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /posts [post]
func (ctrl *PostController) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	// 从 JWT 中获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	post, err := ctrl.postService.Create(req.Title, req.Content, userID, req.CategoryID, req.Tags)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章创建成功", post)
}

// GetPost 获取文章详情
// @Summary      获取文章详情
// @Description  根据ID获取文章详情
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Param        id path int true "文章ID"
// @Success      200  {object}  utils.Response{data=SwaggerPostResponse} "获取成功"
// @Failure      404  {object}  utils.Response "文章不存在"
// @Router       /posts/{id} [get]
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

// UpdatePost 更新文章
// @Summary      更新文章
// @Description  更新文章内容（需要认证，仅作者或管理员）
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "文章ID"
// @Param        request body UpdatePostRequest true "文章信息"
// @Success      200  {object}  utils.Response{data=SwaggerPostResponse} "更新成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "文章不存在"
// @Router       /posts/{id} [put]
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

	// 更新文章
	updatedPost, err := ctrl.postService.UpdatePost(uint(id), req.Title, req.Content, req.CategoryID, req.Tags)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "文章更新成功", updatedPost)
}

// DeletePost 删除文章
// @Summary      删除文章
// @Description  删除文章（需要认证，仅作者或管理员）
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "文章ID"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "文章不存在"
// @Router       /posts/{id} [delete]
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

// ListPosts 获取文章列表
// @Summary      获取文章列表
// @Description  分页获取文章列表
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Param        status query int false "状态 1:已发布 0:草稿 2:私密" default(1)
// @Param        category_id query int false "分类ID"
// @Success      200  {object}  utils.Response{data=SwaggerPostListData} "获取成功"
// @Router       /posts [get]
func (ctrl *PostController) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "1"))
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 32)

	// 获取用户角色（可选认证）
	role, _ := middleware.GetUserRole(c)

	// 如果不是管理员，只显示已发布的文章
	if role != "admin" {
		status = 1
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

// GetPostsByTag 根据标签获取文章
// @Summary      根据标签获取文章
// @Description  通过标签slug获取该标签下的所有文章
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Param        slug path string true "标签slug"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=SwaggerPostListData} "获取成功"
// @Failure      404  {object}  utils.Response "标签不存在"
// @Router       /posts/tag/{slug} [get]
func (ctrl *PostController) GetPostsByTag(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ValidationError(c, "无效的标签Slug")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	posts, total, err := ctrl.postService.GetPostsByTag(slug, page, pageSize)
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
		"tag_slug":    slug,
	})
}
