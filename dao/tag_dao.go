// dao/tag_dao.go
package dao

import (
	"blog/models"
	"blog/utils"
	"gorm.io/gorm"
)

type TagDAO struct {
	db *gorm.DB
}

func NewTagDAO(db *gorm.DB) *TagDAO {
	return &TagDAO{db: db}
}

// Create 创建标签
func (dao *TagDAO) Create(tag *models.Tag) error {
	if err := dao.db.Create(tag).Error; err != nil {
		if utils.IsDuplicateEntryError(err) {
			return utils.ErrTagExists
		}
		return utils.WrapError(utils.ErrInternalServer, "创建标签失败")
	}
	return nil
}

// GetByID 根据ID获取标签
func (dao *TagDAO) GetByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := dao.db.First(&tag, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrTagNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询标签失败")
	}
	return &tag, nil
}

// GetBySlug 根据Slug获取标签
func (dao *TagDAO) GetBySlug(slug string) (*models.Tag, error) {
	var tag models.Tag
	err := dao.db.Where("slug = ?", slug).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrTagNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询标签失败")
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (dao *TagDAO) GetByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := dao.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrTagNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询标签失败")
	}
	return &tag, nil
}

// Update 更新标签
func (dao *TagDAO) Update(tag *models.Tag) error {
	if err := dao.db.Save(tag).Error; err != nil {
		if utils.IsDuplicateEntryError(err) {
			return utils.ErrTagExists
		}
		return utils.WrapError(utils.ErrInternalServer, "更新标签失败")
	}
	return nil
}

// Delete 删除标签
func (dao *TagDAO) Delete(id uint) error {
	// 先删除关联关系
	if err := dao.db.Where("tag_id = ?", id).Delete(&models.PostTag{}).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除标签关联失败")
	}

	result := dao.db.Delete(&models.Tag{}, id)
	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除标签失败")
	}
	if result.RowsAffected == 0 {
		return utils.ErrTagNotFound
	}
	return nil
}

// List 获取标签列表
func (dao *TagDAO) List(page, pageSize int) ([]models.Tag, int64, error) {
	var tags []models.Tag
	var total int64

	offset := (page - 1) * pageSize

	if err := dao.db.Model(&models.Tag{}).Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询标签总数失败")
	}

	err := dao.db.Offset(offset).Limit(pageSize).
		Order("post_count DESC, created_at DESC").
		Find(&tags).Error
	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询标签列表失败")
	}

	return tags, total, nil
}

// GetAll 获取所有标签（不分页）
func (dao *TagDAO) GetAll() ([]models.Tag, error) {
	var tags []models.Tag
	err := dao.db.Order("post_count DESC, created_at DESC").Find(&tags).Error
	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询所有标签失败")
	}
	return tags, nil
}

// GetPopular 获取热门标签（按文章数量排序）
func (dao *TagDAO) GetPopular(limit int) ([]models.Tag, error) {
	var tags []models.Tag
	err := dao.db.Order("post_count DESC, created_at DESC").
		Limit(limit).
		Find(&tags).Error
	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询热门标签失败")
	}
	return tags, nil
}

// CheckNameExists 检查标签名称是否存在
func (dao *TagDAO) CheckNameExists(name string, excludeID ...uint) (bool, error) {
	var count int64
	query := dao.db.Model(&models.Tag{}).Where("name = ?", name)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查标签名称失败")
	}
	return count > 0, nil
}

// CheckSlugExists 检查标签Slug是否存在
func (dao *TagDAO) CheckSlugExists(slug string, excludeID ...uint) (bool, error) {
	var count int64
	query := dao.db.Model(&models.Tag{}).Where("slug = ?", slug)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查标签Slug失败")
	}
	return count > 0, nil
}

// UpdatePostCount 更新标签下的文章数量
func (dao *TagDAO) UpdatePostCount(tagID uint) error {
	var count int64
	if err := dao.db.Model(&models.PostTag{}).Where("tag_id = ?", tagID).Count(&count).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "统计文章数量失败")
	}

	err := dao.db.Model(&models.Tag{}).Where("id = ?", tagID).
		Update("post_count", count).Error
	if err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新标签文章数量失败")
	}
	return nil
}

// AddPostTag 添加文章标签关联
func (dao *TagDAO) AddPostTag(postID, tagID uint) error {
	postTag := &models.PostTag{
		PostID: postID,
		TagID:  tagID,
	}

	// 使用 FirstOrCreate 避免重复
	result := dao.db.Where("post_id = ? AND tag_id = ?", postID, tagID).
		FirstOrCreate(postTag)

	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "添加文章标签关联失败")
	}

	// 更新标签文章数量
	return dao.UpdatePostCount(tagID)
}

// RemovePostTag 删除文章标签关联
func (dao *TagDAO) RemovePostTag(postID, tagID uint) error {
	result := dao.db.Where("post_id = ? AND tag_id = ?", postID, tagID).
		Delete(&models.PostTag{})

	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除文章标签关联失败")
	}

	// 更新标签文章数量
	return dao.UpdatePostCount(tagID)
}

// GetPostTags 获取文章的所有标签
func (dao *TagDAO) GetPostTags(postID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := dao.db.Table("tags").
		Joins("INNER JOIN post_tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id = ?", postID).
		Find(&tags).Error

	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询文章标签失败")
	}
	return tags, nil
}

// BatchAddPostTags 批量添加文章标签
func (dao *TagDAO) BatchAddPostTags(postID uint, tagIDs []uint) error {
	// 使用事务
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// 先删除所有旧关联
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostTag{}).Error; err != nil {
			return err
		}

		// 添加新关联
		for _, tagID := range tagIDs {
			postTag := &models.PostTag{
				PostID: postID,
				TagID:  tagID,
			}
			if err := tx.Create(postTag).Error; err != nil {
				return err
			}
		}

		// 更新所有相关标签的文章数量
		for _, tagID := range tagIDs {
			if err := dao.UpdatePostCount(tagID); err != nil {
				return err
			}
		}

		return nil
	})
}
