
package auth
 
import (
	"errors"
	"time"
 
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)
 
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
 
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}
 
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
 
	signed, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
 
	return signed, nil
}
 
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
 
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
 
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
 
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
 
	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, err
	}
 
	return userID, nil
}

