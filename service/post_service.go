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
	postDAO *dao.PostDAO
}

func NewPostService(postDAO *dao.PostDAO) *PostService {
	return &PostService{postDAO: postDAO}
}

func (s *PostService) Create(title, content string, userID, categoryID uint, tagNames []string) (*models.Post, error) {
	slug := generateSlug(title)

	// 检查slug是否已存在
	_, err := s.postDAO.GetBySlug(slug)
	if err == nil {
		// slug已存在，添加随机后缀
		slug = slug + "-" + utils.GenerateRandomString(6)
	} else if err != utils.ErrPostNotFound {
		return nil, err
	}

	post := &models.Post{
		Title:      title,
		Slug:       slug,
		Content:    content,
		UserID:     userID,
		CategoryID: categoryID,
		Status:     1,
	}

	// 使用事务创建文章和标签
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 创建文章
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// 处理标签
		if len(tagNames) > 0 {
			// 这里需要 tagDAO，可以通过依赖注入
			// 暂时先不处理，后续完善
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetPostByID(id uint) (*models.Post, error) {
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 增加阅读量（异步处理）
	go func() {
		s.postDAO.IncrementViewCount(id)
	}()

	return post, nil
}

func (s *PostService) UpdatePost(id uint, title, content string, categoryID uint) error {
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return err
	}

	post.Title = title
	post.Content = content
	post.CategoryID = categoryID

	return s.postDAO.Update(post)
}

func (s *PostService) DeletePost(id uint) error {
	return s.postDAO.Delete(id)
}

func (s *PostService) ListPosts(page, pageSize int, status int, categoryID uint) ([]models.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.postDAO.List(page, pageSize, status, categoryID)
}

func generateSlug(title string) string {
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

	// 限制长度
	if len(slug) > 200 {
		slug = slug[:200]
	}

	return slug
}

func (s *PostService) UpdatePostCategory(postID, newCategoryID uint) error {
	// 获取原文章
	post, err := s.postDAO.GetByID(postID)
	if err != nil {
		return err
	}

	oldCategoryID := post.CategoryID

	// 更新文章分类
	post.CategoryID = newCategoryID
	if err := s.postDAO.Update(post); err != nil {
		return err
	}

	// 更新旧分类的文章数量
	if oldCategoryID > 0 {
		// 这里需要 categoryDAO，可以在 PostService 中添加 categoryDAO 依赖
	}

	// 更新新分类的文章数量
	if newCategoryID > 0 {
		// 这里需要 categoryDAO，可以在 PostService 中添加 categoryDAO 依赖
	}

	return nil
}
