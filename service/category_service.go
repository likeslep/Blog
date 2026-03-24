// service/category_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
	"strings"
)

type CategoryService struct {
	categoryDAO *dao.CategoryDAO
}

func NewCategoryService(categoryDAO *dao.CategoryDAO) *CategoryService {
	return &CategoryService{categoryDAO: categoryDAO}
}

// Create 创建分类
func (s *CategoryService) Create(name, description string, sortOrder int) (*models.Category, error) {
	// 验证名称
	if utils.IsEmpty(name) {
		return nil, utils.ErrBadRequest
	}

	// 检查名称是否已存在
	exists, err := s.categoryDAO.CheckNameExists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrCategoryExists
	}

	// 生成 slug
	slug := s.generateSlug(name)

	// 确保 slug 唯一
	slug, err = s.ensureUniqueSlug(slug)
	if err != nil {
		return nil, err
	}

	category := &models.Category{
		Name:        name,
		Slug:        slug,
		Description: description,
		SortOrder:   sortOrder,
		PostCount:   0,
	}

	err = s.categoryDAO.Create(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// Update 更新分类
func (s *CategoryService) Update(id uint, name, description string, sortOrder int) (*models.Category, error) {
	category, err := s.categoryDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 如果名称改变了，检查是否重复
	if name != category.Name {
		exists, err := s.categoryDAO.CheckNameExists(name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, utils.ErrCategoryExists
		}

		// 更新 slug
		category.Name = name
		category.Slug = s.generateSlug(name)

		// 确保 slug 唯一
		slug, err := s.ensureUniqueSlug(category.Slug, id)
		if err != nil {
			return nil, err
		}
		category.Slug = slug
	}

	category.Description = description
	category.SortOrder = sortOrder

	err = s.categoryDAO.Update(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// Delete 删除分类
func (s *CategoryService) Delete(id uint) error {
	return s.categoryDAO.Delete(id)
}

// GetByID 根据ID获取分类
func (s *CategoryService) GetByID(id uint) (*models.Category, error) {
	return s.categoryDAO.GetByID(id)
}

// GetBySlug 根据Slug获取分类
func (s *CategoryService) GetBySlug(slug string) (*models.Category, error) {
	return s.categoryDAO.GetBySlug(slug)
}

// List 获取分类列表
func (s *CategoryService) List(page, pageSize int) ([]models.Category, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.categoryDAO.List(page, pageSize)
}

// GetAll 获取所有分类（不分页）
func (s *CategoryService) GetAll() ([]models.Category, error) {
	return s.categoryDAO.GetAll()
}

// generateSlug 生成 slug
func (s *CategoryService) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "/", "")
	slug = strings.ReplaceAll(slug, "\\", "")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, "*", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, "<", "")
	slug = strings.ReplaceAll(slug, ">", "")
	slug = strings.ReplaceAll(slug, "|", "")

	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

// ensureUniqueSlug 确保 slug 唯一
func (s *CategoryService) ensureUniqueSlug(slug string, excludeID ...uint) (string, error) {
	uniqueSlug := slug
	counter := 1

	for {
		exists, err := s.categoryDAO.CheckSlugExists(uniqueSlug, excludeID...)
		if err != nil {
			return "", err
		}
		if !exists {
			break
		}
		uniqueSlug = slug + "-" + utils.GenerateRandomString(4)
		counter++
		if counter > 10 {
			uniqueSlug = slug + "-" + utils.GenerateRandomString(8)
			break
		}
	}

	return uniqueSlug, nil
}
