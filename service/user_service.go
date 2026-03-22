// service/user_service.go
package service

import (
	"blog/dao"
	"blog/models"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userDAO *dao.UserDAO
}

func NewUserService(userDAO *dao.UserDAO) *UserService {
	return &UserService{userDAO: userDAO}
}

func (s *UserService) Register(username, email, password string) (*models.User, error) {
	// 检查用户名是否存在
	existingUser, _ := s.userDAO.GetByUsername(username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	existingEmail, _ := s.userDAO.GetByEmail(email)
	if existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
		Status:   1,
	}

	err = s.userDAO.Create(user)
	return user, err
}

func (s *UserService) Login(username, password string) (*models.User, error) {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}
