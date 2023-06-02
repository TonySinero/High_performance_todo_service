package service

import (
	"errors"
	"fmt"
	"newFeatures/models"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidToken = errors.New("token is invalid")

const (
	AccessTokenTTL  = time.Minute * 60
	RefreshTokenTTL = time.Hour * 24 * 30
)

type MyClaims struct {
	jwt.RegisteredClaims
	UserId   int
	UserRole string
}

func (a *AuthorizationService) GenerateTokens(user *models.User) (*models.GenerateTokens, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &MyClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		user.Id,
		string(user.Role),
	})
	var tokens models.GenerateTokens
	tokenString, err := token.SignedString([]byte(os.Getenv("TOKEN_KEY")))
	if err != nil {
		return nil, err
	}
	tokens.AccessToken = tokenString
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, MyClaims{
		UserId:           user.Id,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL))},
	})
	tokens.RefreshToken, err = refreshToken.SignedString([]byte(os.Getenv("TOKEN_KEY")))
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (a *AuthorizationService) ParseToken(token string) (int, string, error) {
	parseToken, err := jwt.ParseWithClaims(token, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_KEY")), nil
	})
	if err != nil {
		return 0, "", err
	}

	if claims, ok := parseToken.Claims.(*MyClaims); ok && parseToken.Valid {
		return claims.UserId, claims.UserRole, nil
	} else {
		return 0, "", ErrInvalidToken
	}

}
func (a *AuthorizationService) RefreshToken(refreshToken string) (*models.GenerateTokens, error) {
	parseToken, err := jwt.ParseWithClaims(refreshToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parseToken.Claims.(*MyClaims); ok && parseToken.Valid {
		user, err := a.repository.AuthorizationApp.UserRoleById(claims.UserId)
		if err != nil {
			return nil, err
		}
		return a.GenerateTokens(user)
	}

	return nil, errors.New("error parsing token")
}

func (a *AuthorizationService) CheckRole(neededRoles []string, givenRole string) error {
	neededRolesString := strings.Join(neededRoles, ",")
	if !strings.Contains(neededRolesString, givenRole) {
		return fmt.Errorf("not enough rights")
	}
	return nil
}

func ParseTokenGraph(token string) (int, string, error) {
	parseToken, err := jwt.ParseWithClaims(token, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_KEY")), nil
	})
	if err != nil {
		return 0, "", err
	}
	if claims, ok := parseToken.Claims.(*MyClaims); ok && parseToken.Valid {
		return claims.UserId, claims.UserRole, nil
	} else {
		return 0, "", ErrInvalidToken
	}
}
