// dao/user_dao.go
package dao

import (
	"blog/models"
	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (dao *UserDAO) Create(user *models.User) error {
	return dao.db.Create(user).Error
}

func (dao *UserDAO) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := dao.db.First(&user, id).Error
	return &user, err
}

func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (dao *UserDAO) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (dao *UserDAO) Update(user *models.User) error {
	return dao.db.Save(user).Error
}

func (dao *UserDAO) Delete(id uint) error {
	return dao.db.Delete(&models.User{}, id).Error
}

func (dao *UserDAO) List(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * pageSize

	if err := dao.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dao.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	return users, total, err
}
