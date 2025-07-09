package pkg

import (
	"time"
	"os"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(uid string) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Hour * 72).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		uid, _ := claims["uid"].(string)
		return uid, nil
	}

	return "", errors.New("invalid claims")
}
