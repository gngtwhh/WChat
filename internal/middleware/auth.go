package middleware

import (
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

// Auth JWT 鉴权中间件
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, errcode.TokenMissing)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, errcode.TokenInvalid, "Token 格式错误，应为 Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]
		user, err := authSvc.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			response.FailErr(c, err)
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, user.Uuid)
		c.Set(ContextIsAdminKey, user.IsAdmin)
		c.Set(ContextRawTokenKey, tokenString)

		c.Next()
	}
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
