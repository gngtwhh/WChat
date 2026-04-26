package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"wchat/internal/service"
	"wchat/pkg/errcode"
	"wchat/pkg/response"
)

const (
	ContextUserIDKey   = "user_uuid"
	ContextIsAdminKey  = "is_admin"
	ContextRawTokenKey = "raw_token"
)

// Auth validates REST requests with Authorization: Bearer <token>.
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := tokenFromAuthorizationHeader(c)
		if err != nil {
			response.Fail(c, errcode.TokenInvalid, err.Error())
			c.Abort()
			return
		}
		if tokenString == "" {
			response.Fail(c, errcode.TokenMissing)
			c.Abort()
			return
		}

		if !authenticateWithToken(c, authSvc, tokenString) {
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthWS validates websocket upgrades with Authorization first, then query token.
func AuthWS(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := tokenFromAuthorizationHeader(c)
		if err != nil {
			response.Fail(c, errcode.TokenInvalid, err.Error())
			c.Abort()
			return
		}
		if tokenString == "" {
			tokenString = strings.TrimSpace(c.Query("token"))
		}
		if tokenString == "" {
			response.Fail(c, errcode.TokenMissing)
			c.Abort()
			return
		}

		if !authenticateWithToken(c, authSvc, tokenString) {
			c.Abort()
			return
		}

		c.Next()
	}
}

func tokenFromAuthorizationHeader(c *gin.Context) (string, error) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return "", nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("%s,Format: Bearer <token>", errcode.GetMsg(errcode.TokenInvalid))
	}

	return strings.TrimSpace(parts[1]), nil
}

func authenticateWithToken(c *gin.Context, authSvc *service.AuthService, tokenString string) bool {
	user, err := authSvc.ValidateToken(c.Request.Context(), tokenString)
	if err != nil {
		response.FailErr(c, err)
		return false
	}

	c.Set(ContextUserIDKey, user.Uuid)
	c.Set(ContextIsAdminKey, user.IsAdmin)
	c.Set(ContextRawTokenKey, tokenString)
	return true
}

func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return "", false
	}
	uuid, ok := val.(string)
	return uuid, ok
}

func CheckAdmin(c *gin.Context) bool {
	val, exists := c.Get(ContextIsAdminKey)
	if !exists {
		return false
	}
	isAdmin, ok := val.(int8)
	return ok && isAdmin == 1
}

func GetRawToken(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextRawTokenKey)
	if !exists {
		return "", false
	}
	token, ok := val.(string)
	return token, ok
}
