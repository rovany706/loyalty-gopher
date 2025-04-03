package auth

import "time"

const tokenExpiryTime = time.Hour

type TokenManager interface {
	GetClaimsFromToken(tokenString string) (*Claims, error)
	CreateToken(userID int) (string, error)
}
