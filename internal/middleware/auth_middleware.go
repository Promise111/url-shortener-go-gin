package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authHeader string = c.GetHeader("authorization")

		var tokenString = strings.TrimPrefix(authHeader, "bearer ")
		if authHeader == tokenString || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "message": "Invalid authorization header format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid token claims",
			})
			return
		}

		userId, ok := claims["id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Invalid token claims",
			})
			return
		}

		if exp, ok := claims["exp"].(float64); !ok {
			var expirationTime = time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status":  false,
					"message": "Token expired",
				})
				return
			}
		}

		c.Set("user_id", userId)
		c.Next()
	}
}
