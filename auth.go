package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key - in production, use environment variable
var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
	GuestName string `json:"guest_name"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	Token     string `json:"token"`
	GuestName string `json:"guest_name"`
	ExpiresAt int64  `json:"expires_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// generateGuestToken creates a JWT token for a guest user
func generateGuestToken(guestName string) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		GuestName: guestName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

// validateToken validates a JWT token and returns the claims
func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// handleGetToken generates and returns a guest token
func handleGetToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Generate unique guest name
	guestName := fmt.Sprintf("guest-%s", randomHexStrings())

	// Generate JWT token
	token, expiresAt, err := generateGuestToken(guestName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Return token response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		Token:     token,
		GuestName: guestName,
		ExpiresAt: expiresAt,
	})
}

// extractTokenFromRequest extracts JWT token from request
// Supports: Authorization header, query parameter, or first message
func extractTokenFromRequest(r *http.Request) (string, error) {
	// Method 1: Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// Method 2: Check query parameter
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no token found in request")
}

// authenticateWebSocket validates the token and returns guest name
func authenticateWebSocket(r *http.Request) (string, error) {
	token, err := extractTokenFromRequest(r)
	if err != nil {
		return "", err
	}

	claims, err := validateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	return claims.GuestName, nil
}
