// dao/attachment_dao.go
package dao

import (
	"blog/models"
	"blog/utils"
	"gorm.io/gorm"
)

type AttachmentDAO struct {
	db *gorm.DB
}

func NewAttachmentDAO(db *gorm.DB) *AttachmentDAO {
	return &AttachmentDAO{db: db}
}

// Create 创建附件记录
func (dao *AttachmentDAO) Create(attachment *models.Attachment) error {
	if err := dao.db.Create(attachment).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "创建附件记录失败")
	}
	return nil
}

// GetByID 根据ID获取附件
func (dao *AttachmentDAO) GetByID(id uint) (*models.Attachment, error) {
	var attachment models.Attachment
	err := dao.db.First(&attachment, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrAttachmentNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询附件失败")
	}
	return &attachment, nil
}

// Delete 删除附件记录
func (dao *AttachmentDAO) Delete(id uint) error {
	result := dao.db.Delete(&models.Attachment{}, id)
	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除附件记录失败")
	}
	if result.RowsAffected == 0 {
		return utils.ErrAttachmentNotFound
	}
	return nil
}

// List 获取附件列表
func (dao *AttachmentDAO) List(page, pageSize int, fileType string, userID uint) ([]models.Attachment, int64, error) {
	var attachments []models.Attachment
	var total int64

	offset := (page - 1) * pageSize
	query := dao.db.Model(&models.Attachment{})

	if fileType != "" {
		query = query.Where("type = ?", fileType)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询附件总数失败")
	}

	err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&attachments).Error

	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询附件列表失败")
	}

	return attachments, total, nil
}

// GetUserAttachments 获取用户的所有附件
func (dao *AttachmentDAO) GetUserAttachments(userID uint) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := dao.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&attachments).Error

	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "查询用户附件失败")
	}

	return attachments, nil
}
