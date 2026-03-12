package middleware

import (
    "strings"

    "github.com/gin-gonic/gin"

    "wchat/pkg/errcode"
    "wchat/pkg/response"
    "wchat/pkg/utils"
)

const (
    ContextUserIDKey  = "user_uuid"
    ContextIsAdminKey = "is_admin"
)

// Auth JWT 鉴权中间件
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Authorization: Bearer <token>
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Fail(c, errcode.TokenMissing)
            c.Abort() // 拦截请求
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Fail(c, errcode.TokenInvalid, "Token 格式错误，应为 Bearer <token>")
            c.Abort()
            return
        }

        tokenString := parts[1]

        claims, err := utils.ParseToken(tokenString)
        if err != nil {
            response.Fail(c, errcode.TokenExpired, "Token 失效或已过期，请重新登录")
            c.Abort()
            return
        }

        c.Set(ContextUserIDKey, claims.Uuid)
        c.Set(ContextIsAdminKey, claims.IsAdmin)

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
