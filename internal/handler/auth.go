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

// Register 用户注册
// @Summary      用户注册
// @Description  通过手机号、密码和昵称注册新账号
// @Tags         鉴权
// @Accept       json
// @Produce      json
// @Param        req  body      types.RegisterReq                true  "注册参数"
// @Success      200  {object}  response.Response{data=nil}      "注册成功"
// @Failure      200  {object}  response.Response{data=nil}      "参数错误 / 手机号已注册"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req types.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}
	// TODO: 手机验证码/图形验证码
	err := h.authSvc.Register(c.Request.Context(), req.Telephone, req.Password, req.Nickname)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.Success(c, nil)
}

// Login 用户登录
// @Summary      用户登录
// @Description  通过手机号和密码登录，验证成功后下发 JWT Token 及用户基本信息
// @Tags         鉴权
// @Accept       json
// @Produce      json
// @Param        req  body      types.LoginReq                        true  "登录参数"
// @Success      200  {object}  response.Response{data=types.LoginResp}  "登录成功，返回 token 和 user_info"
// @Failure      200  {object}  response.Response{data=nil}              "账号或密码错误 / 账号被禁用"
// @Router       /auth/login [post]
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

// Logout 用户登出
// @Summary      用户登出
// @Description  登出当前账号，更新最后离线时间（后续可扩展 Token 黑名单机制）
// @Tags         鉴权
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=nil}  "登出成功"
// @Failure      200  {object}  response.Response{data=nil}  "Token 无效"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	token, _ := middleware.GetRawToken(c)

	if err := h.authSvc.Logout(c.Request.Context(), userID, token); err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}
