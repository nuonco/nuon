package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

type IDClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

type UserInfo struct {
	Email string
	Name  string
}

func (a *Service) getUserInfo(IDToken string) UserInfo {
	claims := IDClaims{}
	jwt.ParseWithClaims(IDToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})

	user := UserInfo{
		Email: claims.Email,
		Name:  claims.Name,
	}

	return user
}
