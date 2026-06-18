package service

import (
	"errors"
	"fmt"
	"incubator-backend/internal/config"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService interface {
	Login(username, password string) (string, *models.User, error)
	ParseToken(tokenString string) (*CustomClaims, error)
	GetUserByID(id uint) (*models.User, error)
	InitAdminUser() error
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Login(username, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		logger.Errorf("get user by username failed: %v", err)
		return "", nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return "", nil, errors.New("用户已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := s.generateToken(user)
	if err != nil {
		logger.Errorf("generate token failed: %v", err)
		return "", nil, errors.New("生成令牌失败")
	}

	now := time.Now()
	user.LastLogin = &now
	_ = s.userRepo.Update(user)

	return token, user, nil
}

func (s *authService) generateToken(user *models.User) (string, error) {
	expireHours := config.GlobalConfig.JWT.ExpireHours
	claims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "incubator-backend",
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
}

func (s *authService) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GlobalConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *authService) InitAdminUser() error {
	_, err := s.userRepo.GetByUsername("admin")
	if err == nil {
		logger.Info("admin user already exists")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "系统管理员",
		Role:     "admin",
		Status:   1,
	}

	err = s.userRepo.Create(admin)
	if err != nil {
		logger.Errorf("create admin user failed: %v", err)
		return err
	}

	logger.Info("admin user created successfully")
	return nil
}
