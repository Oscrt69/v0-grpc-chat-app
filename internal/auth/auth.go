package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	JWTSecret     = "your-secret-key-change-this-in-production"
	TokenDuration = 24 * time.Hour
)

// GenerateToken membuat JWT token baru
func GenerateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(TokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken memvalidasi JWT token
func ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		jwt.MapClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateID membuat UUID baru
func GenerateID() string {
	return uuid.New().String()
}

// HashPassword melakukan hashing password (simplified - gunakan bcrypt di production)
func HashPassword(password string) string {
	// Di production, gunakan library bcrypt
	// import "golang.org/x/crypto/bcrypt"
	// hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return password // Simplified untuk demo
}

// VerifyPassword memverifikasi password (simplified)
func VerifyPassword(hashedPassword, password string) bool {
	// Di production, gunakan:
	// return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
	return hashedPassword == password // Simplified untuk demo
}

// GetUserIDFromToken mengekstrak user_id dari token
func GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}

	return userID, nil
}

// GetUsernameFromToken mengekstrak username dari token
func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username not found in token")
	}

	return username, nil
}
