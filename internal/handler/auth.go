package handler

import (
    "wchat/internal/types"

    "github.com/gin-gonic/gin"

    "wchat/internal/middleware"
    "wchat/internal/service"
    "wchat/pkg/errcode"
    "wchat/pkg/response"
)

type AuthHandler struct {
    authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
    return &AuthHandler{authSvc: authSvc}
}

// ==============================================================
// 1. 注册接口 (Register)
// ==============================================================

// Register 处理用户账号注册请求
//
// Route:   POST /api/v1/auth/register
// Auth:    公开接口
// Request: JSON -> RegisterReq
// Returns: JSON -> { code: 0, msg: "ok", data: null }
func (h *AuthHandler) Register(c *gin.Context) {
    var req types.RegisterReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "参数格式错误")
        return
    }
    err := h.authSvc.Register(c.Request.Context(), req.Telephone, req.Password, req.Nickname)
    if err != nil {
        response.FailErr(c, err)
    }
    response.Success(c, nil)
}

// ==============================================================
// 2. 登录接口 (Login)
// ==============================================================

// Login 处理用户登录请求，验证并下发 JWT
//
// Route:   POST /api/v1/auth/login
// Auth:    公开接口
// Request: JSON -> LoginReq
// Returns: JSON -> { code: 0, msg: "ok", data: LoginResp }
func (h *AuthHandler) Login(c *gin.Context) {
    var req types.LoginReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "参数格式错误")
        return
    }

    user, token, err := h.authSvc.Login(c.Request.Context(), req.Telephone, req.Password)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    resp := types.LoginResp{
        Token: token,
        UserInfo: types.UserVO{
            Uuid:      user.Uuid,
            Nickname:  user.Nickname,
            Avatar:    user.Avatar,
            Signature: user.Signature,
        },
    }

    response.Success(c, resp)
}

// ==============================================================
// 3. 注销接口 (Logout)
// ==============================================================

// Logout 处理用户登出请求
//
// Route:   POST /api/v1/auth/logout
// Auth:    需要 Token 鉴权 (通过 middleware.Auth())
// Request: 无 Body
// Returns: JSON -> { code: 0, msg: "ok", data: null }
func (h *AuthHandler) Logout(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    if err := h.authSvc.Logout(c.Request.Context(), userID); err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}
