// dao/category_dao.go
package dao

import (
	"blog/models"
	"blog/utils"
	"gorm.io/gorm"
)

type CategoryDAO struct {
	db *gorm.DB
}

func NewCategoryDAO(db *gorm.DB) *CategoryDAO {
	return &CategoryDAO{db: db}
}

// Create 创建分类
func (dao *CategoryDAO) Create(category *models.Category) error {
	if err := dao.db.Create(category).Error; err != nil {
		if utils.IsDuplicateEntryError(err) {
			return utils.ErrCategoryExists
		}
		return utils.WrapError(utils.ErrInternalServer, "创建分类失败")
	}
	return nil
}

// GetByID 根据ID获取分类
func (dao *CategoryDAO) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := dao.db.First(&category, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询分类失败")
	}
	return &category, nil
}

// GetBySlug 根据Slug获取分类
func (dao *CategoryDAO) GetBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := dao.db.Where("slug = ?", slug).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询分类失败")
	}
	return &category, nil
}

// GetByName 根据名称获取分类
func (dao *CategoryDAO) GetByName(name string) (*models.Category, error) {
	var category models.Category
	err := dao.db.Where("name = ?", name).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询分类失败")
	}
	return &category, nil
}

// Update 更新分类
func (dao *CategoryDAO) Update(category *models.Category) error {
	if err := dao.db.Save(category).Error; err != nil {
		if utils.IsDuplicateEntryError(err) {
			return utils.ErrCategoryExists
		}
		return utils.WrapError(utils.ErrInternalServer, "更新分类失败")
	}
	return nil
}

// Delete 删除分类
func (dao *CategoryDAO) Delete(id uint) error {
	// 先检查是否有文章使用这个分类
	var count int64
	if err := dao.db.Model(&models.Post{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "检查分类使用情况失败")
	}

	if count > 0 {
		return utils.ErrCategoryHasPosts
	}

	result := dao.db.Delete(&models.Category{}, id)
	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除分类失败")
	}
	if result.RowsAffected == 0 {
		return utils.ErrCategoryNotFound
	}
	return nil
}

// List 获取分类列表
func (dao *CategoryDAO) List(page, pageSize int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64

	offset := (page - 1) * pageSize

	if err := dao.db.Model(&models.Category{}).Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询分类总数失败")
	}

	err := dao.db.Offset(offset).Limit(pageSize).
		Order("sort_order ASC, created_at DESC").
		Find(&categories).Error
	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询分类列表失败")
	}

	return categories, total, nil
}

// GetAll 获取所有分类（不分页）
func (dao *CategoryDAO) GetAll() ([]models.Category, error) {
	var categories []models.Category
	err := dao.db.Order("sort_order ASC, created_at DESC").Find(&categories).Error
	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询所有分类失败")
	}
	return categories, nil
}

// UpdatePostCount 更新分类下的文章数量
func (dao *CategoryDAO) UpdatePostCount(categoryID uint) error {
	var count int64
	if err := dao.db.Model(&models.Post{}).Where("category_id = ? AND status = ?", categoryID, 1).Count(&count).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "统计文章数量失败")
	}

	err := dao.db.Model(&models.Category{}).Where("id = ?", categoryID).
		Update("post_count", count).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新分类文章数量失败")
	}
	return nil
}

// CheckNameExists 检查分类名称是否存在
func (dao *CategoryDAO) CheckNameExists(name string, excludeID ...uint) (bool, error) {
	var count int64
	query := dao.db.Model(&models.Category{}).Where("name = ?", name)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查分类名称失败")
	}
	return count > 0, nil
}

// CheckSlugExists 检查分类Slug是否存在
func (dao *CategoryDAO) CheckSlugExists(slug string, excludeID ...uint) (bool, error) {
	var count int64
	query := dao.db.Model(&models.Category{}).Where("slug = ?", slug)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查分类Slug失败")
	}
	return count > 0, nil
}
