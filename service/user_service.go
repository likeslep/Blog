// service/user_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"blog/utils"
)

type UserService struct {
	userDAO *dao.UserDAO
}

func NewUserService(userDAO *dao.UserDAO) *UserService {
	return &UserService{userDAO: userDAO}
}

func (s *UserService) Register(username, email, password string) (*models.User, error) {
	// 检查用户名是否存在
	exists, err := s.userDAO.IsUsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrUserAlreadyExists
	}

	// 检查邮箱是否存在
	exists, err = s.userDAO.IsEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrEmailAlreadyExists
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, utils.WrapError(utils.ErrInternalServer, "密码加密失败")
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
		Status:   1,
	}

	err = s.userDAO.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(username, password string) (*models.User, error) {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, utils.ErrInvalidPassword
		}
		return nil, err
	}

	if user.Status != 1 {
		return nil, utils.ErrUserDisabled
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, utils.ErrInvalidPassword
	}

	return user, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userDAO.GetByID(id)
}

func (s *UserService) UpdateUser(user *models.User) error {
	return s.userDAO.Update(user)
}
