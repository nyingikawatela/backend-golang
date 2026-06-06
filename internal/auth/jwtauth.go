
package auth
 
import (
	"errors"
	"time"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	 "crypto/rand"
    "encoding/hex"
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

func GetBearerToken(headers http.Header) (string, error){
	authHeader := headers.Get("Authorization")
	if authHeader == ""{
		return "", errors.New("Token inexistente")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func MakeRefreshToken() string{
	bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return ""
    }
    return hex.EncodeToString(bytes)
}