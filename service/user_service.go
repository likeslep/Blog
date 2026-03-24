// service/user_service.go
package service

import (
	"blog/config"
	"blog/dao"
	"blog/models"
	"blog/utils"
	"time"
)

type UserService struct {
	userDAO   *dao.UserDAO
	jwtConfig utils.JWTConfig
}

func NewUserService(userDAO *dao.UserDAO, cfg *config.Config) *UserService {
	jwtConfig := utils.JWTConfig{
		Secret:     cfg.JWT.Secret,
		ExpireTime: time.Duration(cfg.JWT.ExpireHour) * time.Hour,
		Issuer:     cfg.JWT.Issuer,
	}
	return &UserService{
		userDAO:   userDAO,
		jwtConfig: jwtConfig,
	}
}

func (s *UserService) Register(username, email, password string) (*models.User, string, error) {
	// 检查用户名是否存在
	exists, err := s.userDAO.IsUsernameExists(username)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", utils.ErrUserAlreadyExists
	}

	// 检查邮箱是否存在
	exists, err = s.userDAO.IsEmailExists(email)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", utils.ErrEmailAlreadyExists
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", utils.WrapError(utils.ErrInternalServer, "密码加密失败")
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
		return nil, "", err
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtConfig)
	if err != nil {
		return nil, "", utils.WrapError(utils.ErrInternalServer, "生成令牌失败")
	}

	return user, token, nil
}

func (s *UserService) Login(username, password string) (*models.User, string, error) {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, "", utils.ErrInvalidPassword
		}
		return nil, "", err
	}

	if user.Status != 1 {
		return nil, "", utils.ErrUserDisabled
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", utils.ErrInvalidPassword
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	s.userDAO.Update(user)

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtConfig)
	if err != nil {
		return nil, "", utils.WrapError(utils.ErrInternalServer, "生成令牌失败")
	}

	return user, token, nil
}

// 刷新 token
func (s *UserService) RefreshToken(oldToken string) (string, error) {
	return utils.RefreshToken(oldToken, s.jwtConfig)
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userDAO.GetByID(id)
}

func (s *UserService) UpdateUser(user *models.User) error {
	return s.userDAO.Update(user)
}
