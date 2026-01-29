package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	params := argon2id.Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(password, &params)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Printf("Error comparing password and hash: %v\n", err)
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	expires := now.Add(expiresIn)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "p4t",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
		Subject:   userID.String(),
	})
	signingKey := []byte(tokenSecret)
	tokenString, err := token.SignedString(signingKey)

	if err != nil {
		fmt.Printf("Error signing jwt: %v\n", err)
		return "", err
	}
	return tokenString, nil
}

// Validate returns the id associated with the JWT - it is still up to the handler to verify
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("token validation failed %v", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("Token is invalid")
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Could not parse subject into a uuid: %s", claims.Subject)
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authValues := headers["Authorization"]
	if len(authValues) == 0 {
		return "", fmt.Errorf("No values in the authorization header")
	}
	parts := strings.Fields(authValues[0])
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("Missing or improperly formatted/headed token: %s", authValues[0])
	}
	token := parts[1]
	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("Error creating a random number for refresh token: %v", err)
	}
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apiKeyValues := headers["Authorization"]
	if len(apiKeyValues) == 0 {
		return "", fmt.Errorf("No values in the authorization header")
	}

	parts := strings.Fields(apiKeyValues[0])
	if len(parts) != 2 {
		return "", fmt.Errorf("Malformed Authorization field in header %v", apiKeyValues)
	}
	apiKey := parts[1]
	return apiKey, nil
}
