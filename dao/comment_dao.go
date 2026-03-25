// dao/comment_dao.go
package dao

import (
	"blog/models"
	"blog/utils"

	"gorm.io/gorm"
)

type CommentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) *CommentDAO {
	return &CommentDAO{db: db}
}

// Create 创建评论
func (dao *CommentDAO) Create(comment *models.Comment) error {
	if err := dao.db.Create(comment).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "创建评论失败")
	}
	return nil
}

// GetByID 根据ID获取评论
func (dao *CommentDAO) GetByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := dao.db.Preload("User").First(&comment, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrCommentNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询评论失败")
	}
	return &comment, nil
}

// Update 更新评论
func (dao *CommentDAO) Update(comment *models.Comment) error {
	if err := dao.db.Save(comment).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新评论失败")
	}
	return nil
}

// Delete 删除评论
func (dao *CommentDAO) Delete(id uint) error {
	// 使用事务删除评论及其所有子评论
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// 递归删除所有子评论
		if err := tx.Where("parent_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		// 删除评论本身
		result := tx.Delete(&models.Comment{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return utils.ErrCommentNotFound
		}
		return nil
	})
}

// GetPostComments 获取文章的所有顶级评论（分页）
func (dao *CommentDAO) GetPostComments(postID uint, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	// 查询顶级评论（parent_id IS NULL）
	query := dao.db.Model(&models.Comment{}).
		Where("post_id = ? AND parent_id IS NULL AND status = ?", postID, 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询评论总数失败")
	}

	err := query.Preload("User").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询评论列表失败")
	}

	// 加载每个评论的子回复
	for i := range comments {
		replies, err := dao.GetCommentReplies(comments[i].ID)
		if err != nil {
			return nil, 0, err
		}
		comments[i].Replies = replies
	}

	return comments, total, nil
}

// GetCommentReplies 获取评论的所有回复
func (dao *CommentDAO) GetCommentReplies(commentID uint) ([]models.Comment, error) {
	var replies []models.Comment
	err := dao.db.Preload("User").
		Where("parent_id = ? AND status = ?", commentID, 1).
		Order("created_at ASC").
		Find(&replies).Error

	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询回复失败")
	}

	return replies, nil
}

// GetAllCommentsByPost 获取文章的所有评论（不分页，用于管理）
func (dao *CommentDAO) GetAllCommentsByPost(postID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := dao.db.Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询评论失败")
	}

	return comments, nil
}

// GetUserComments 获取用户的所有评论
func (dao *CommentDAO) GetUserComments(userID uint, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	query := dao.db.Model(&models.Comment{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询用户评论总数失败")
	}

	err := query.Preload("Post").Preload("User").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询用户评论失败")
	}

	return comments, total, nil
}

// IncrementLikeCount 增加点赞数
func (dao *CommentDAO) IncrementLikeCount(id uint) error {
	err := dao.db.Model(&models.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新点赞数失败")
	}
	return nil
}

// DecrementLikeCount 减少点赞数
func (dao *CommentDAO) DecrementLikeCount(id uint) error {
	err := dao.db.Model(&models.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新点赞数失败")
	}
	return nil
}

// UpdateStatus 更新评论状态
func (dao *CommentDAO) UpdateStatus(id uint, status int) error {
	err := dao.db.Model(&models.Comment{}).Where("id = ?", id).
		Update("status", status).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新评论状态失败")
	}
	return nil
}

// GetPendingComments 获取待审核评论
func (dao *CommentDAO) GetPendingComments(page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	query := dao.db.Model(&models.Comment{}).Where("status = ?", 0)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询待审核评论总数失败")
	}

	err := query.Preload("User").Preload("Post").
		Offset(offset).Limit(pageSize).
		Order("created_at ASC").
		Find(&comments).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询待审核评论失败")
	}

	return comments, total, nil
}
