// service/post_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
	"strings"

	"gorm.io/gorm"
)

type PostService struct {
	postDAO     *dao.PostDAO
	tagDAO      *dao.TagDAO
	categoryDAO *dao.CategoryDAO
	db          *gorm.DB
}

func NewPostService(postDAO *dao.PostDAO, tagDAO *dao.TagDAO, categoryDAO *dao.CategoryDAO, db *gorm.DB) *PostService {
	return &PostService{
		postDAO:     postDAO,
		tagDAO:      tagDAO,
		categoryDAO: categoryDAO,
		db:          db,
	}
}

// GetPostByID 根据ID获取文章
func (s *PostService) GetPostByID(id uint) (*models.Post, error) {
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 增加阅读量（异步）
	go func() {
		s.postDAO.IncrementViewCount(id)
	}()

	return post, nil
}

// ListPosts 获取文章列表
func (s *PostService) ListPosts(page, pageSize int, status int, categoryID uint) ([]models.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.postDAO.List(page, pageSize, status, categoryID)
}

// Create 创建文章
func (s *PostService) Create(title, content string, userID, categoryID uint, tagNames []string) (*models.Post, error) {
	// 验证标题
	if utils.IsEmpty(title) {
		return nil, utils.ErrBadRequest
	}

	// 生成 slug
	slug := s.generateSlug(title)

	// 确保 slug 唯一
	slug, err := s.ensureUniqueSlug(slug)
	if err != nil {
		return nil, err
	}

	// 处理分类（如果提供了分类ID）
	if categoryID > 0 {
		_, err := s.categoryDAO.GetByID(categoryID)
		if err != nil {
			return nil, err
		}
	}

	// 处理标签
	var tagIDs []uint
	if len(tagNames) > 0 {
		tagIDs, err = s.processTags(tagNames)
		if err != nil {
			return nil, err
		}
	}

	// 创建文章
	post := &models.Post{
		Title:      title,
		Slug:       slug,
		Content:    content,
		UserID:     userID,
		CategoryID: categoryID,
		Status:     1,
	}

	// 使用事务创建文章和标签关联
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建文章
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// 添加标签关联
		if len(tagIDs) > 0 {
			for _, tagID := range tagIDs {
				if err := s.tagDAO.AddPostTag(post.ID, tagID); err != nil {
					return err
				}
			}
		}

		// 更新分类文章数量
		if categoryID > 0 {
			if err := s.categoryDAO.UpdatePostCount(categoryID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新加载文章（包含关联数据）
	return s.postDAO.GetByID(post.ID)
}

// UpdatePost 更新文章
func (s *PostService) UpdatePost(id uint, title, content string, categoryID uint, tagNames []string) (*models.Post, error) {
	// 获取原文章
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldCategoryID := post.CategoryID

	// 更新基本信息
	post.Title = title
	post.Content = content

	// 处理分类变更
	if categoryID != post.CategoryID {
		// 验证新分类是否存在
		if categoryID > 0 {
			_, err := s.categoryDAO.GetByID(categoryID)
			if err != nil {
				return nil, err
			}
		}
		post.CategoryID = categoryID
	}

	// 处理标签
	var tagIDs []uint
	if len(tagNames) > 0 {
		tagIDs, err = s.processTags(tagNames)
		if err != nil {
			return nil, err
		}
	}

	// 使用事务更新文章和标签
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新文章
		if err := tx.Save(post).Error; err != nil {
			return err
		}

		// 更新标签关联
		if len(tagIDs) > 0 {
			// 批量更新标签关联
			if err := s.tagDAO.BatchAddPostTags(post.ID, tagIDs); err != nil {
				return err
			}
		} else {
			// 如果没有标签，删除所有关联
			if err := tx.Where("post_id = ?", post.ID).Delete(&models.PostTag{}).Error; err != nil {
				return err
			}
		}

		// 更新分类文章数量
		if oldCategoryID > 0 {
			if err := s.categoryDAO.UpdatePostCount(oldCategoryID); err != nil {
				return err
			}
		}
		if post.CategoryID > 0 && post.CategoryID != oldCategoryID {
			if err := s.categoryDAO.UpdatePostCount(post.CategoryID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新加载文章（包含关联数据）
	return s.postDAO.GetByID(post.ID)
}

// DeletePost 删除文章
func (s *PostService) DeletePost(id uint) error {
	// 获取文章信息
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return err
	}

	// 使用事务删除文章
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除文章
		if err := tx.Delete(&models.Post{}, id).Error; err != nil {
			return err
		}

		// 更新分类文章数量
		if post.CategoryID > 0 {
			if err := s.categoryDAO.UpdatePostCount(post.CategoryID); err != nil {
				return err
			}
		}

		// 标签关联会自动级联删除（如果设置了外键约束）
		// 但为了安全，手动删除
		if err := tx.Where("post_id = ?", id).Delete(&models.PostTag{}).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// GetPostsByTag 根据标签获取文章列表
func (s *PostService) GetPostsByTag(tagSlug string, page, pageSize int) ([]models.Post, int64, error) {
	// 获取标签
	tag, err := s.tagDAO.GetBySlug(tagSlug)
	if err != nil {
		return nil, 0, err
	}

	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	// 查询该标签下的文章
	query := s.db.Model(&models.Post{}).
		Joins("INNER JOIN post_tags ON post_tags.post_id = posts.id").
		Where("post_tags.tag_id = ? AND posts.status = ?", tag.ID, 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询文章总数失败")
	}

	err = query.Preload("User").Preload("Category").
		Offset(offset).Limit(pageSize).
		Order("posts.is_top DESC, posts.created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询文章列表失败")
	}

	return posts, total, nil
}

// processTags 处理标签列表，返回标签ID列表
func (s *PostService) processTags(tagNames []string) ([]uint, error) {
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
				tag, err = s.createTag(name)
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

// createTag 创建标签
func (s *PostService) createTag(name string) (*models.Tag, error) {
	slug := s.generateTagSlug(name)

	// 确保 slug 唯一
	slug, err := s.ensureUniqueTagSlug(slug)
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

// generateSlug 生成文章slug
func (s *PostService) generateSlug(title string) string {
	slug := strings.ToLower(title)
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

	if len(slug) > 200 {
		slug = slug[:200]
	}

	return slug
}

// ensureUniqueSlug 确保文章slug唯一
func (s *PostService) ensureUniqueSlug(slug string) (string, error) {
	uniqueSlug := slug
	counter := 1

	for {
		_, err := s.postDAO.GetBySlug(uniqueSlug)
		if err != nil {
			if err == utils.ErrPostNotFound {
				break
			}
			return "", err
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

// generateTagSlug 生成标签slug
func (s *PostService) generateTagSlug(name string) string {
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

// ensureUniqueTagSlug 确保标签slug唯一
func (s *PostService) ensureUniqueTagSlug(slug string) (string, error) {
	uniqueSlug := slug
	counter := 1

	for {
		_, err := s.tagDAO.GetBySlug(uniqueSlug)
		if err != nil {
			if err == utils.ErrTagNotFound {
				break
			}
			return "", err
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
