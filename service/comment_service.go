// service/comment_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"

	"gorm.io/gorm"
)

type CommentService struct {
	commentDAO *dao.CommentDAO
	postDAO    *dao.PostDAO
	userDAO    *dao.UserDAO
	db         *gorm.DB
}

func NewCommentService(commentDAO *dao.CommentDAO, postDAO *dao.PostDAO, userDAO *dao.UserDAO, db *gorm.DB) *CommentService {
	return &CommentService{
		commentDAO: commentDAO,
		postDAO:    postDAO,
		userDAO:    userDAO,
		db:         db,
	}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(postID, userID uint, content string, parentID *uint) (*models.Comment, error) {
	// 验证内容
	if utils.IsEmpty(content) {
		return nil, utils.ErrBadRequest
	}

	// 验证文章是否存在
	_, err := s.postDAO.GetByID(postID)
	if err != nil {
		return nil, err
	}

	// 验证用户是否存在
	_, err = s.userDAO.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 如果 parentID 不为空，验证父评论是否存在
	if parentID != nil && *parentID > 0 {
		parentComment, err := s.commentDAO.GetByID(*parentID)
		if err != nil {
			return nil, err
		}
		// 确保父评论属于同一篇文章
		if parentComment.PostID != postID {
			return nil, utils.ErrBadRequest
		}
	}

	// 创建评论
	comment := &models.Comment{
		Content:  content,
		UserID:   userID,
		PostID:   postID,
		ParentID: parentID,
		Status:   1, // 默认直接发布，可以改为0需要审核
	}

	err = s.commentDAO.Create(comment)
	if err != nil {
		return nil, err
	}

	// 更新文章的评论数
	go func() {
		s.postDAO.IncrementCommentCount(postID)
	}()

	// 重新加载评论（包含用户信息）
	return s.commentDAO.GetByID(comment.ID)
}

// GetPostComments 获取文章评论
func (s *CommentService) GetPostComments(postID uint, page, pageSize int) ([]models.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.commentDAO.GetPostComments(postID, page, pageSize)
}

// GetUserComments 获取用户评论
func (s *CommentService) GetUserComments(userID uint, page, pageSize int) ([]models.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.commentDAO.GetUserComments(userID, page, pageSize)
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(commentID, userID uint, role string) error {
	comment, err := s.commentDAO.GetByID(commentID)
	if err != nil {
		return err
	}

	// 权限检查：只有评论作者或管理员可以删除
	if comment.UserID != userID && role != "admin" {
		return utils.ErrInsufficientPermission
	}

	// 删除评论
	err = s.commentDAO.Delete(commentID)
	if err != nil {
		return err
	}

	// 更新文章的评论数（重新计算）
	go func() {
		var count int64
		s.db.Model(&models.Comment{}).Where("post_id = ?", comment.PostID).Count(&count)
		s.postDAO.UpdateCommentCount(comment.PostID, int(count))
	}()

	return nil
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(commentID uint) error {
	_, err := s.commentDAO.GetByID(commentID)
	if err != nil {
		return err
	}

	return s.commentDAO.IncrementLikeCount(commentID)
}

// UnlikeComment 取消点赞
func (s *CommentService) UnlikeComment(commentID uint) error {
	comment, err := s.commentDAO.GetByID(commentID)
	if err != nil {
		return err
	}

	if comment.LikeCount <= 0 {
		return nil
	}

	return s.commentDAO.DecrementLikeCount(commentID)
}

// ApproveComment 审核通过评论（管理员）
func (s *CommentService) ApproveComment(commentID uint) error {
	comment, err := s.commentDAO.GetByID(commentID)
	if err != nil {
		return err
	}

	if comment.Status == 1 {
		return nil // 已经是通过状态
	}

	err = s.commentDAO.UpdateStatus(commentID, 1)
	if err != nil {
		return err
	}

	// 更新文章的评论数
	go func() {
		s.postDAO.IncrementCommentCount(comment.PostID)
	}()

	return nil
}

// RejectComment 拒绝评论（管理员）
func (s *CommentService) RejectComment(commentID uint) error {
	_, err := s.commentDAO.GetByID(commentID)
	if err != nil {
		return err
	}

	return s.commentDAO.UpdateStatus(commentID, 2)
}

// GetPendingComments 获取待审核评论（管理员）
func (s *CommentService) GetPendingComments(page, pageSize int) ([]models.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.commentDAO.GetPendingComments(page, pageSize)
}
