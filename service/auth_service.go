package service

import (
	"context"
	"errors"
	"fmt"
	"newFeatures/mail"
	"newFeatures/models"
	"newFeatures/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthorizationService struct {
	repository *repository.Repository
}

var (
	ErrEmptyFields          = errors.New("the fields are empty")
	ErrInvalidPassOrName    = errors.New("password or name less then 6 symbols")
	ErrDeletedUser          = errors.New("this user is deleted")
	ErrIncorrectCredentials = errors.New("incorrect phone number or password")
	ErrUserNotFound         = errors.New("user does not exist")
	ErrorEmailDoesNotExist  = errors.New("user with this email does not exist")
)

func (a *AuthorizationService) CreateUser(ctx context.Context, user *models.User) (*models.GenerateTokens, error) {
	if err := validateUser(user); err != nil {
		return nil, err
	}
	var password = user.Password
	hash, err := HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = hash
	err = a.repository.AuthorizationApp.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	go mail.SendEmail(&models.Post{
		Email:    user.Email,
		Password: password,
	})
	token, err := a.GenerateTokens(user)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (a *AuthorizationService) AuthUser(ctx context.Context, user *models.User) (tokens *models.GenerateTokens, err error) {
	userDB, err := a.repository.AuthorizationApp.UserByPhone(ctx, user)
	if err != nil {
		return nil, err
	}
	if !CheckPasswordHash(user.Password, userDB.Password) {
		return nil, ErrIncorrectCredentials
	}

	tokens, err = a.GenerateTokens(userDB)
	if err != nil {
		return nil, err
	}
	return tokens, nil

}

func (a *AuthorizationService) User(ctx context.Context, userID int) (*models.ResponseUser, error) {
	user, err := a.repository.AuthorizationApp.UserById(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *AuthorizationService) Users(ctx context.Context, page, limit int64) ([]models.ResponseUser, error) {
	users, err := a.repository.AuthorizationApp.Users(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *AuthorizationService) UpdateUser(ctx context.Context, inputUser *models.ResponseUser) error {
	userDB, err := a.repository.AuthorizationApp.UserById(ctx, inputUser.Id)
	if err != nil {
		return ErrUserNotFound
	}

	if inputUser.Name == "" {
		inputUser.Name = userDB.Name
	}
	if inputUser.Email == "" {
		inputUser.Email = userDB.Email
	}
	if inputUser.Phone == "" {
		inputUser.Phone = userDB.Phone
	}

	return a.repository.AuthorizationApp.UpdateUser(ctx, inputUser)
}

func (a *AuthorizationService) DeleteUser(ctx context.Context, id int) error {
	err := a.repository.AuthorizationApp.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthorizationService) RestorePassword(ctx context.Context, restore *models.RestorePassword) error {
	err := a.repository.AuthorizationApp.CheckByEmail(ctx, restore)
	if err != nil {
		return err
	}
	hash, err := HashPassword(restore.Password)
	if err != nil {
		return fmt.Errorf("RestorePassword: can not generate hash from password:%w", err)
	}
	restore.Password = hash
	err = a.repository.AuthorizationApp.RestorePassword(ctx, restore)
	if err != nil {
		return err
	}
	go mail.SendEmail(&models.Post{
		Email:    restore.Email,
		Password: restore.Password,
	})
	return nil
}

func validateUser(user *models.User) error {
	if user.Name == "" || user.Email == "" || user.Phone == "" || user.Password == "" {
		return ErrEmptyFields
	}
	if len(user.Password) < 6 || len(user.Name) < 6 {
		return ErrInvalidPassOrName
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
