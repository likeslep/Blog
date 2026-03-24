// service/tag_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
	"strings"
)

type TagService struct {
	tagDAO *dao.TagDAO
}

func NewTagService(tagDAO *dao.TagDAO) *TagService {
	return &TagService{tagDAO: tagDAO}
}

// Create 创建标签
func (s *TagService) Create(name, description string) (*models.Tag, error) {
	// 验证名称
	if utils.IsEmpty(name) {
		return nil, utils.ErrBadRequest
	}

	// 检查名称是否已存在
	exists, err := s.tagDAO.CheckNameExists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrTagExists
	}

	// 生成 slug
	slug := s.generateSlug(name)

	// 确保 slug 唯一
	slug, err = s.ensureUniqueSlug(slug)
	if err != nil {
		return nil, err
	}

	tag := &models.Tag{
		Name:      name,
		Slug:      slug,
		PostCount: 0,
	}

	err = s.tagDAO.Create(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Update 更新标签
func (s *TagService) Update(id uint, name string) (*models.Tag, error) {
	tag, err := s.tagDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 如果名称改变了，检查是否重复
	if name != tag.Name {
		exists, err := s.tagDAO.CheckNameExists(name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, utils.ErrTagExists
		}

		tag.Name = name
		tag.Slug = s.generateSlug(name)

		// 确保 slug 唯一
		slug, err := s.ensureUniqueSlug(tag.Slug, id)
		if err != nil {
			return nil, err
		}
		tag.Slug = slug
	}

	err = s.tagDAO.Update(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Delete 删除标签
func (s *TagService) Delete(id uint) error {
	return s.tagDAO.Delete(id)
}

// GetByID 根据ID获取标签
func (s *TagService) GetByID(id uint) (*models.Tag, error) {
	return s.tagDAO.GetByID(id)
}

// GetBySlug 根据Slug获取标签
func (s *TagService) GetBySlug(slug string) (*models.Tag, error) {
	return s.tagDAO.GetBySlug(slug)
}

// List 获取标签列表
func (s *TagService) List(page, pageSize int) ([]models.Tag, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.tagDAO.List(page, pageSize)
}

// GetAll 获取所有标签
func (s *TagService) GetAll() ([]models.Tag, error) {
	return s.tagDAO.GetAll()
}

// GetPopular 获取热门标签
func (s *TagService) GetPopular(limit int) ([]models.Tag, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	return s.tagDAO.GetPopular(limit)
}

// GetPostTags 获取文章的所有标签
func (s *TagService) GetPostTags(postID uint) ([]models.Tag, error) {
	return s.tagDAO.GetPostTags(postID)
}

// ProcessTags 处理标签列表（创建不存在的标签，返回ID列表）
func (s *TagService) ProcessTags(tagNames []string) ([]uint, error) {
	var tagIDs []uint

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// 查找是否已存在
		tag, err := s.tagDAO.GetByName(name)
		if err != nil {
			if err == utils.ErrTagNotFound {
				// 不存在，创建新标签
				tag, err = s.Create(name, "")
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		tagIDs = append(tagIDs, tag.ID)
	}

	return tagIDs, nil
}

// generateSlug 生成 slug
func (s *TagService) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "#", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "/", "")
	slug = strings.ReplaceAll(slug, "\\", "")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, "*", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, "<", "")
	slug = strings.ReplaceAll(slug, ">", "")
	slug = strings.ReplaceAll(slug, "|", "")
	slug = strings.ReplaceAll(slug, "&", "")

	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

// ensureUniqueSlug 确保 slug 唯一
func (s *TagService) ensureUniqueSlug(slug string, excludeID ...uint) (string, error) {
	uniqueSlug := slug
	counter := 1

	for {
		exists, err := s.tagDAO.CheckSlugExists(uniqueSlug, excludeID...)
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
