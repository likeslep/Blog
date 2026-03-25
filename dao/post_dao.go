// dao/post_dao.go
package dao

import (
	"blog/models"
	"blog/utils"

	"gorm.io/gorm"
)

type PostDAO struct {
	db *gorm.DB
}

func NewPostDAO(db *gorm.DB) *PostDAO {
	return &PostDAO{db: db}
}

func (dao *PostDAO) Create(post *models.Post) error {
	if err := dao.db.Create(post).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "创建文章失败")
	}
	return nil
}

func (dao *PostDAO) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	err := dao.db.Preload("User").Preload("Category").Preload("Tags").First(&post, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrPostNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询文章失败")
	}
	return &post, nil
}

func (dao *PostDAO) GetBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := dao.db.Preload("User").Preload("Category").Preload("Tags").
		Where("slug = ?", slug).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrPostNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询文章失败")
	}
	return &post, nil
}

func (dao *PostDAO) Update(post *models.Post) error {
	if err := dao.db.Save(post).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新文章失败")
	}
	return nil
}

func (dao *PostDAO) Delete(id uint) error {
	result := dao.db.Delete(&models.Post{}, id)
	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除文章失败")
	}
	if result.RowsAffected == 0 {
		return utils.ErrPostNotFound
	}
	return nil
}

func (dao *PostDAO) List(page, pageSize int, status int, categoryID uint) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize
	query := dao.db.Model(&models.Post{})

	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询文章总数失败")
	}

	err := query.Preload("User").Preload("Category").
		Offset(offset).Limit(pageSize).
		Order("is_top DESC, created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询文章列表失败")
	}

	return posts, total, nil
}

func (dao *PostDAO) IncrementViewCount(id uint) error {
	err := dao.db.Model(&models.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新阅读量失败")
	}
	return nil
}

func (dao *PostDAO) IncrementCommentCount(id uint) error {
	err := dao.db.Model(&models.Post{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新评论数失败")
	}
	return nil
}

// UpdateCommentCount 更新文章评论数
func (dao *PostDAO) UpdateCommentCount(postID uint, count int) error {
	err := dao.db.Model(&models.Post{}).Where("id = ?", postID).
		Update("comment_count", count).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新评论数失败")
	}
	return nil
}
