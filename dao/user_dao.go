// dao/user_dao.go
package dao

import (
	"blog/models"
	"blog/utils"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (dao *UserDAO) Create(user *models.User) error {
	if err := dao.db.Create(user).Error; err != nil {
		// 处理唯一约束冲突
		if utils.IsDuplicateEntryError(err) {
			if err.Error() == "Duplicate entry" {
				return utils.ErrUserAlreadyExists
			}
		}
		return utils.WrapError(utils.ErrInternalServer, "创建用户失败")
	}
	return nil
}

func (dao *UserDAO) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := dao.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询用户失败")
	}
	return &user, nil
}

func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询用户失败")
	}
	return &user, nil
}

func (dao *UserDAO) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrUserNotFound
		}
		return nil, utils.WrapError(utils.ErrInternalServer, "查询用户失败")
	}
	return &user, nil
}

func (dao *UserDAO) Update(user *models.User) error {
	if err := dao.db.Save(user).Error; err != nil {
		return utils.WrapError(utils.ErrInternalServer, "更新用户失败")
	}
	return nil
}

func (dao *UserDAO) Delete(id uint) error {
	result := dao.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return utils.WrapError(utils.ErrInternalServer, "删除用户失败")
	}
	if result.RowsAffected == 0 {
		return utils.ErrUserNotFound
	}
	return nil
}

func (dao *UserDAO) List(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * pageSize

	if err := dao.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询用户总数失败")
	}

	err := dao.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, utils.WrapError(utils.ErrInternalServer, "查询用户列表失败")
	}

	return users, total, nil
}

// 检查用户名是否存在
func (dao *UserDAO) IsUsernameExists(username string) (bool, error) {
	var count int64
	err := dao.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查用户名失败")
	}
	return count > 0, nil
}

// 检查邮箱是否存在
func (dao *UserDAO) IsEmailExists(email string) (bool, error) {
	var count int64
	err := dao.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, utils.WrapError(utils.ErrInternalServer, "检查邮箱失败")
	}
	return count > 0, nil
}
