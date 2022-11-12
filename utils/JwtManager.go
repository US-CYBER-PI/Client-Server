package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt"
)

type JwtManager struct {
	hmacSecret []byte
}

func NewJwtManager(hmacSecret []byte) *JwtManager {
	return &JwtManager{
		hmacSecret: hmacSecret,
	}
}

func (j *JwtManager) GetTokenClaim(token string) *jwt.MapClaims {

	tokenR, _ := j.GetToken(token)

	if tokenR == nil {
		return nil
	}

	claims, _ := tokenR.Claims.(jwt.MapClaims)

	return &claims
}

func (j *JwtManager) GetToken(token string) (*jwt.Token, string) {

	tokenResult, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return j.hmacSecret, nil
	})

	if err != nil {
		return nil, ""
	}

	tokenSha := Sha512(token)

	if _, ok := tokenResult.Claims.(jwt.MapClaims); ok && tokenResult.Valid {
		return tokenResult, tokenSha
	}

	return nil, ""
}
