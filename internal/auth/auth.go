package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const secretKey = "supersecretkey"
const AuthCookieName = "Authorization"

// claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

func SetNewAuthCookie(userID uuid.UUID, res http.ResponseWriter) error {
	jwtString, err := buildJWTString(userID)
	if err != nil {
		zap.L().Sugar().Debugln("Error building jwtString", err.Error())
		return err
	}

	authCookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    jwtString,
		HttpOnly: true, // Prevents JavaScript access
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   3600, // Cookie expires in 1 hour
	}

	http.SetCookie(res, authCookie)

	return nil
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func buildJWTString(userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string) (userID uuid.UUID, err error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
