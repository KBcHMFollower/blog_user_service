package tokens_helper

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type TokenClaims struct {
	Email string
	Id    uuid.UUID
}

func CreateNewJwt(userId uuid.UUID, email string, tokenTTL time.Duration, tokenSecret string) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userId
	claims["email"] = email
	claims["exp"] = time.Now().Add(tokenTTL).Unix()
	tokenString, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", fmt.Errorf("error in creating  jwt proccess: %v", err)
	}

	return tokenString, nil
}

func Parse(tokenString string, tokenSecret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unaviable sign-metod: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if token.Valid {
		return token, nil
	}
	return nil, err
}

func GetClaimsValues(token *jwt.Token) (TokenClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return TokenClaims{}, fmt.Errorf("can`t parse claims")
	}

	userIdString, ok := claims["user_id"].(string)
	if !ok {
		return TokenClaims{}, fmt.Errorf("can`t convert  user_id")
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return TokenClaims{}, fmt.Errorf("%w", err)
	}

	userEmail, ok := claims["user_id"].(string)
	if !ok {
		return TokenClaims{}, fmt.Errorf("%w", err)
	}

	return TokenClaims{
		Email: userEmail,
		Id:    userId,
	}, nil
}
