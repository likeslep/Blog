// controller/comment_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SwaggerCommentResponse 评论响应（用于 Swagger 文档）
type SwaggerCommentResponse struct {
	ID        uint                     `json:"id" example:"1"`
	Content   string                   `json:"content" example:"写得很好！"`
	UserID    uint                     `json:"user_id" example:"1"`
	Username  string                   `json:"username" example:"testuser"`
	PostID    uint                     `json:"post_id" example:"1"`
	ParentID  *uint                    `json:"parent_id"` // 移除 example 标签
	LikeCount int                      `json:"like_count" example:"0"`
	Status    int                      `json:"status" example:"1"`
	CreatedAt string                   `json:"created_at" example:"2024-01-01T00:00:00Z"`
	Replies   []SwaggerCommentResponse `json:"replies,omitempty"`
}

// SwaggerCommentListData 评论列表响应（用于 Swagger 文档）
type SwaggerCommentListData struct {
	List       []SwaggerCommentResponse `json:"list"`
	Total      int64                    `json:"total" example:"100"`
	Page       int                      `json:"page" example:"1"`
	PageSize   int                      `json:"page_size" example:"10"`
	TotalPages int64                    `json:"total_pages" example:"10"`
}

type CommentController struct {
	commentService *service.CommentService
}

func NewCommentController(commentService *service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required,min=1,max=1000" example:"写得很好！"`
	ParentID *uint  `json:"parent_id"`
}

// CreateComment 发表评论
// @Summary      发表评论
// @Description  对文章发表评论或回复（需要认证）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "文章ID"
// @Param        request body CreateCommentRequest true "评论内容"
// @Success      200  {object}  utils.Response{data=SwaggerCommentResponse} "评论成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      404  {object}  utils.Response "文章不存在"
// @Router       /posts/{id}/comments [post]
func (ctrl *CommentController) CreateComment(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	comment, err := ctrl.commentService.CreateComment(uint(postID), userID, req.Content, req.ParentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "评论成功", comment)
}

// GetPostComments 获取文章评论列表
// @Summary      获取文章评论列表
// @Description  获取指定文章的所有评论（树形结构）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Param        id path int true "文章ID"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=SwaggerCommentListData} "获取成功"
// @Router       /posts/{id}/comments [get]
func (ctrl *CommentController) GetPostComments(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的文章ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := ctrl.commentService.GetPostComments(uint(postID), page, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        comments,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetUserComments 获取用户评论列表
// @Summary      获取用户评论列表
// @Description  获取当前登录用户的所有评论（需要认证）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=SwaggerCommentListData} "获取成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /profile/comments [get]
func (ctrl *CommentController) GetUserComments(c *gin.Context) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := ctrl.commentService.GetUserComments(userID, page, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        comments,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// DeleteComment 删除评论
// @Summary      删除评论
// @Description  删除评论（需要认证，仅作者或管理员）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cid path int true "评论ID"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "评论不存在"
// @Router       /comments/{cid} [delete]
func (ctrl *CommentController) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的评论ID")
		return
	}

	// 获取当前用户信息
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	role, _ := middleware.GetUserRole(c)

	err = ctrl.commentService.DeleteComment(uint(commentID), userID, role)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "评论删除成功", nil)
}

// LikeComment 点赞评论
// @Summary      点赞评论
// @Description  给评论点赞（需要认证）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cid path int true "评论ID"
// @Success      200  {object}  utils.Response "点赞成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      404  {object}  utils.Response "评论不存在"
// @Router       /comments/{cid}/like [post]
func (ctrl *CommentController) LikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的评论ID")
		return
	}

	err = ctrl.commentService.LikeComment(uint(commentID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "点赞成功", nil)
}

// UnlikeComment 取消点赞
// @Summary      取消点赞
// @Description  取消评论点赞（需要认证）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cid path int true "评论ID"
// @Success      200  {object}  utils.Response "取消点赞成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      404  {object}  utils.Response "评论不存在"
// @Router       /comments/{cid}/like [delete]
func (ctrl *CommentController) UnlikeComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的评论ID")
		return
	}

	err = ctrl.commentService.UnlikeComment(uint(commentID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "取消点赞成功", nil)
}

// 管理员接口

// GetPendingComments 获取待审核评论
// @Summary      获取待审核评论
// @Description  获取所有待审核的评论（需要管理员权限）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=SwaggerCommentListData} "获取成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Router       /admin/comments/pending [get]
func (ctrl *CommentController) GetPendingComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := ctrl.commentService.GetPendingComments(page, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        comments,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ApproveComment 审核通过评论
// @Summary      审核通过评论
// @Description  审核通过待审核的评论（需要管理员权限）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cid path int true "评论ID"
// @Success      200  {object}  utils.Response "审核通过"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "评论不存在"
// @Router       /admin/comments/{cid}/approve [post]
func (ctrl *CommentController) ApproveComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的评论ID")
		return
	}

	err = ctrl.commentService.ApproveComment(uint(commentID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "审核通过", nil)
}

// RejectComment 拒绝评论
// @Summary      拒绝评论
// @Description  拒绝待审核的评论（需要管理员权限）
// @Tags         评论管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cid path int true "评论ID"
// @Success      200  {object}  utils.Response "已拒绝"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "评论不存在"
// @Router       /admin/comments/{cid}/reject [post]
func (ctrl *CommentController) RejectComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的评论ID")
		return
	}

	err = ctrl.commentService.RejectComment(uint(commentID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "已拒绝", nil)
}
