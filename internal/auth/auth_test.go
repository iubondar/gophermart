package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	// Initialize logger for tests
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func TestSetNewAuthCookie(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the function
	err := SetNewAuthCookie(userID, rr)
	require.NoError(t, err)

	// Get the response cookies
	resp := rr.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	require.Len(t, cookies, 1)

	// Check cookie properties
	cookie := cookies[0]
	assert.Equal(t, AuthCookieName, cookie.Name)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, 3600, cookie.MaxAge)
}

func TestBuildJWTString(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()

	// Call the function
	tokenString, err := buildJWTString(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse the token to verify its contents
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	// Verify claims
	claims, ok := token.Claims.(*claims)
	require.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
}

func TestGetUserIDFromReq(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()

	// Create a valid JWT token
	tokenString, err := buildJWTString(userID)
	require.NoError(t, err)

	// Create a request with the cookie
	req := httptest.NewRequest("GET", "/", nil)
	cookie := &http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	}
	req.AddCookie(cookie)

	// Test valid cookie
	extractedUserID, err := GetUserIDFromReq(req)
	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	// Test missing cookie
	req = httptest.NewRequest("GET", "/", nil)
	_, err = GetUserIDFromReq(req)
	assert.Error(t, err)
}

func TestGetUserID(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()

	// Test valid token
	tokenString, err := buildJWTString(userID)
	require.NoError(t, err)

	extractedUserID, err := getUserID(tokenString)
	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	// Test invalid token
	_, err = getUserID("invalid.token.string")
	assert.Error(t, err)

	// Test expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		UserID: userID,
	})
	expiredTokenString, err := expiredToken.SignedString([]byte(secretKey))
	require.NoError(t, err)
	_, err = getUserID(expiredTokenString)
	assert.Error(t, err)
}
