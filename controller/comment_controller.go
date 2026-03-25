// controller/comment_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	commentService *service.CommentService
}

func NewCommentController(commentService *service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required,min=1,max=1000"`
	ParentID *uint  `json:"parent_id"`
}

// CreateComment 创建评论（需要认证）
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

// GetPostComments 获取文章评论列表（公开）
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

// GetUserComments 获取用户评论列表（需要认证）
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

// DeleteComment 删除评论（需要认证）
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

// LikeComment 点赞评论（需要认证）
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

// UnlikeComment 取消点赞（需要认证）
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

// GetPendingComments 获取待审核评论（管理员）
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

// ApproveComment 审核通过评论（管理员）
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

// RejectComment 拒绝评论（管理员）
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
