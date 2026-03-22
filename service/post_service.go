// service/post_service.go
package service

import (
	"blog/dao"
	"blog/models"
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

	post := &models.Post{
		Title:      title,
		Slug:       slug,
		Content:    content,
		UserID:     userID,
		CategoryID: categoryID,
		Status:     1,
	}

	err := s.postDAO.Create(post)
	return post, err
}

func (s *PostService) GetPostByID(id uint) (*models.Post, error) {
	post, err := s.postDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 增加阅读量
	s.postDAO.IncrementViewCount(id)

	return post, nil
}

func (s *PostService) ListPosts(page, pageSize int, status int, categoryID uint) ([]models.Post, int64, error) {
	return s.postDAO.List(page, pageSize, status, categoryID)
}

func generateSlug(title string) string {
	// 简单的slug生成，实际应用中需要更复杂的处理
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "/", "")
	return slug
}
