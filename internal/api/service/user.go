package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"template/internal/model"
)

type UserService struct {
	db          *gorm.DB
	jwtService  *JwtService
	rbcaService *RBACService
}

func NewAuthService(i do.Injector) (*UserService, error) {
	return &UserService{
		db:          do.MustInvoke[*gorm.DB](i),
		jwtService:  do.MustInvoke[*JwtService](i),
		rbcaService: do.MustInvoke[*RBACService](i),
	}, nil
}

func (s *UserService) Register(ctx context.Context, phone, password string) error {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$|^\+?[1-9]\d{1,14}$`, phone)
	if !matched {
		return errors.New("invalid phone number format")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Phone:        phone,
		PasswordHash: string(hash),
		Balance:      1000.0,
	}

	return s.db.WithContext(ctx).Create(user).Error
}

func (s *UserService) Login(ctx context.Context, phone, password string) (string, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid phone or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid phone or password")
	}

	return s.jwtService.GenerateToken(user.ID)
}
