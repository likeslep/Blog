// service/post_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
	"strings"
)

type PostService struct {
	postDAO *dao.PostDAO
}

func NewPostService(postDAO *dao.PostDAO) *PostService {
	return &PostService{postDAO: postDAO}
}

func (s *PostService) Create(title, content string, userID, categoryID uint, tags []string) (*models.Post, error) {
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

	err = s.postDAO.Create(post)
	if err != nil {
		return nil, err
	}

	// TODO: 处理标签关联

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
