// dao/post_dao.go
package dao

import (
	"blog/models"
	"gorm.io/gorm"
)

type PostDAO struct {
	db *gorm.DB
}

func NewPostDAO(db *gorm.DB) *PostDAO {
	return &PostDAO{db: db}
}

func (dao *PostDAO) Create(post *models.Post) error {
	return dao.db.Create(post).Error
}

func (dao *PostDAO) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	err := dao.db.Preload("User").Preload("Category").Preload("Tags").First(&post, id).Error
	return &post, err
}

func (dao *PostDAO) GetBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := dao.db.Preload("User").Preload("Category").Preload("Tags").
		Where("slug = ?", slug).First(&post).Error
	return &post, err
}

func (dao *PostDAO) Update(post *models.Post) error {
	return dao.db.Save(post).Error
}

func (dao *PostDAO) Delete(id uint) error {
	return dao.db.Delete(&models.Post{}, id).Error
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
		return nil, 0, err
	}

	err := query.Preload("User").Preload("Category").
		Offset(offset).Limit(pageSize).
		Order("is_top DESC, created_at DESC").
		Find(&posts).Error

	return posts, total, err
}

func (dao *PostDAO) IncrementViewCount(id uint) error {
	return dao.db.Model(&models.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
