// service/attachment_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
	"mime/multipart"
	"path/filepath"
)

type AttachmentService struct {
	attachmentDAO *dao.AttachmentDAO
	uploadConfig  utils.UploadConfig
	baseURL       string
}

func NewAttachmentService(attachmentDAO *dao.AttachmentDAO, uploadPath string, baseURL string) *AttachmentService {
	return &AttachmentService{
		attachmentDAO: attachmentDAO,
		uploadConfig: utils.UploadConfig{
			UploadPath:  uploadPath,
			MaxSize:     10 * 1024 * 1024, // 10MB
			AllowTypes:  []string{"image/", "video/", "application/pdf", "text/plain"},
			Thumbnail:   true,
			ThumbWidth:  300,
			ThumbHeight: 200,
		},
		baseURL: baseURL,
	}
}

// Upload 上传文件
func (s *AttachmentService) Upload(fileHeader interface{}, userID uint) (*models.Attachment, error) {
	// 类型断言
	fh, ok := fileHeader.(*multipart.FileHeader)
	if !ok {
		return nil, utils.ErrBadRequest
	}

	// 上传文件
	result, err := utils.UploadFile(fh, s.uploadConfig, userID)
	if err != nil {
		return nil, err
	}

	// 保存记录
	attachment := &models.Attachment{
		UserID:       userID,
		Filename:     result.Filename,
		OriginalName: result.OriginalName,
		Path:         result.Path,
		Size:         result.Size,
		MimeType:     result.MimeType,
		Type:         result.Type,
		Width:        result.Width,
		Height:       result.Height,
		ThumbPath:    result.ThumbPath,
	}

	if err := s.attachmentDAO.Create(attachment); err != nil {
		// 上传成功但保存失败，删除已上传的文件
		utils.DeleteFile(filepath.Join(s.uploadConfig.UploadPath, result.Path))
		return nil, err
	}

	// 设置访问URL
	attachment.URL = utils.GetFileURL(result.Path, s.baseURL)

	return attachment, nil
}

// Delete 删除附件
func (s *AttachmentService) Delete(id uint, userID uint, role string) error {
	attachment, err := s.attachmentDAO.GetByID(id)
	if err != nil {
		return err
	}

	// 权限检查：只有上传者或管理员可以删除
	if attachment.UserID != userID && role != "admin" {
		return utils.ErrInsufficientPermission
	}

	// 删除物理文件
	fullPath := filepath.Join(s.uploadConfig.UploadPath, attachment.Path)
	if err := utils.DeleteFile(fullPath); err != nil {
		return err
	}

	// 删除数据库记录
	return s.attachmentDAO.Delete(id)
}

// GetByID 根据ID获取附件
func (s *AttachmentService) GetByID(id uint) (*models.Attachment, error) {
	attachment, err := s.attachmentDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 设置访问URL
	attachment.URL = utils.GetFileURL(attachment.Path, s.baseURL)
	if attachment.ThumbPath != "" {
		attachment.ThumbURL = utils.GetFileURL(attachment.ThumbPath, s.baseURL)
	}

	return attachment, nil
}

// List 获取附件列表
func (s *AttachmentService) List(page, pageSize int, fileType string, userID uint) ([]models.Attachment, int64, error) {
	attachments, total, err := s.attachmentDAO.List(page, pageSize, fileType, userID)
	if err != nil {
		return nil, 0, err
	}

	// 设置访问URL
	for i := range attachments {
		attachments[i].URL = utils.GetFileURL(attachments[i].Path, s.baseURL)
		if attachments[i].ThumbPath != "" {
			attachments[i].ThumbURL = utils.GetFileURL(attachments[i].ThumbPath, s.baseURL)
		}
	}

	return attachments, total, nil
}

// GetUserAttachments 获取用户的所有附件
func (s *AttachmentService) GetUserAttachments(userID uint) ([]models.Attachment, error) {
	attachments, err := s.attachmentDAO.GetUserAttachments(userID)
	if err != nil {
		return nil, err
	}

	// 设置访问URL
	for i := range attachments {
		attachments[i].URL = utils.GetFileURL(attachments[i].Path, s.baseURL)
		if attachments[i].ThumbPath != "" {
			attachments[i].ThumbURL = utils.GetFileURL(attachments[i].ThumbPath, s.baseURL)
		}
	}

	return attachments, nil
}
