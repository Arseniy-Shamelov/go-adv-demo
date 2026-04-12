package auth

import (
	"errors"
	"go-adv/internal/user"
	"go-adv/pkg/di"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository di.IUserRepository
}

func NewAuthService(userRepository di.IUserRepository) *AuthService {
	return &AuthService{UserRepository: userRepository}
}

func (service *AuthService) Login(email, password string) (string, error) {
	existedUser, _ := service.UserRepository.FindByEmail(email)
	if existedUser == nil {
		return "", errors.New(ErrWrongCredentials)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existedUser.Password), []byte(password)); err != nil {
		return "", errors.New(ErrWrongCredentials)
	}

	return existedUser.Email, nil
}

func (service *AuthService) Register(email, password, name string) (string, error) {
	existedUser, _ := service.UserRepository.FindByEmail(email)
	if existedUser != nil {
		return "", errors.New(ErrUserExists)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &user.User{
		Email:    email,
		Name:     name,
		Password: string(hashedPassword),
	}

	if _, err = service.UserRepository.Create(user); err != nil {
		return "", err
	}

	return user.Email, nil
}
