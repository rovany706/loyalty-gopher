package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenManager struct {
	secretKey []byte
}

func NewJWTTokenManager(secretKey []byte) (*JWTTokenManager, error) {
	var err error
	if secretKey == nil {
		secretKey, err = generateSecretKey()

		if err != nil {
			return nil, err
		}
	}

	return &JWTTokenManager{
		secretKey: secretKey,
	}, nil
}

func generateSecretKey() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (auth *JWTTokenManager) CreateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiryTime)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(auth.secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (auth *JWTTokenManager) GetClaimsFromToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return auth.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
