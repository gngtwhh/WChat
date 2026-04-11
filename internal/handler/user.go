package handler

import (
	"wchat/internal/types"

	"github.com/gin-gonic/gin"

	"wchat/internal/middleware"
	"wchat/internal/service"
	"wchat/pkg/errcode"
	"wchat/pkg/response"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetMyProfile 获取当前用户个人信息
// @Summary      获取我的个人信息
// @Description  根据 Token 中的用户 UUID 返回完整的个人资料
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=types.UserProfileResp}  "个人信息"
// @Failure      200  {object}  response.Response{data=nil}                     "Token 无效 / 用户不存在"
// @Router       /users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	uuid, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	user, err := h.svc.GetUserByUUID(c.Request.Context(), uuid)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, types.BuildUserProfileResp(user))
}

// UpdateMyProfile 更新当前用户个人信息
// @Summary      更新我的个人信息
// @Description  修改当前用户的昵称、邮箱、头像、性别、个性签名等字段，仅传入需要修改的字段
// @Tags         用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.UpdateProfileReq           true  "需要更新的字段"
// @Success      200  {object}  response.Response{data=nil}      "更新成功"
// @Failure      200  {object}  response.Response{data=nil}      "参数错误 / Token 无效"
// @Router       /users/me [put]
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	uuid, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	// filter empty fields
	updateData := make(map[string]any)
	if req.Nickname != "" {
		updateData["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}
	if req.Gender != nil {
		updateData["gender"] = *req.Gender
	}
	if req.Signature != "" {
		updateData["signature"] = req.Signature
	}

	err := h.svc.UpdateUserProfile(c.Request.Context(), uuid, updateData)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// RequestAccountDeletion 发起账号注销申请
// @Summary      发起账号注销申请
// @Description  当前登录用户提交注销申请，校验密码后进入注销冷静期，并使当前登录态失效
// @Tags         用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.RequestAccountDeletionReq  true  "注销申请参数"
// @Success      200  {object}  response.Response{data=nil}      "申请成功"
// @Failure      200  {object}  response.Response{data=nil}      "参数错误 / 密码错误 / Token 无效"
// @Router       /users/me/delete-request [post]
func (h *UserHandler) RequestAccountDeletion(c *gin.Context) {
	uuid, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.RequestAccountDeletionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	token, _ := middleware.GetRawToken(c)
	if err := h.svc.RequestAccountDeletion(c.Request.Context(), uuid, req.Password, token); err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// CancelAccountDeletion 取消账号注销申请
// @Summary      取消账号注销申请
// @Description  当前登录用户在冷静期内校验密码后取消注销申请，恢复账号正常状态
// @Tags         用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.CancelAccountDeletionReq  true  "取消注销参数"
// @Success      200  {object}  response.Response{data=nil}     "取消成功"
// @Failure      200  {object}  response.Response{data=nil}     "参数错误 / 密码错误 / Token 无效"
// @Router       /users/me/delete-request [delete]
func (h *UserHandler) CancelAccountDeletion(c *gin.Context) {
	uuid, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.CancelAccountDeletionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	if err := h.svc.CancelAccountDeletion(c.Request.Context(), uuid, req.Password); err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// ChangeTelephone 更换手机号
// @Summary      更换手机号
// @Description  当前登录用户校验密码和验证码后，将登录手机号迁移到新的手机号
// @Tags         用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.ChangeTelephoneReq  true  "更换手机号参数"
// @Success      200  {object}  response.Response{data=nil}  "更换成功"
// @Failure      200  {object}  response.Response{data=nil}  "参数错误 / 密码错误 / 手机号已存在"
// @Router       /users/me/phone/change [post]
func (h *UserHandler) ChangeTelephone(c *gin.Context) {
	uuid, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.ChangeTelephoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	if err := h.svc.ChangeTelephone(c.Request.Context(), uuid, req.Password, req.NewTelephone, req.VerifyCode); err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUserList 搜索用户列表
// @Summary      搜索用户 / 获取用户列表
// @Description  分页获取用户列表，支持通过昵称或手机号进行模糊搜索
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Param        page     query    int     true   "页码 (从1开始)"   minimum(1)
// @Param        size     query    int     true   "每页数量"         minimum(1) maximum(100)
// @Param        keyword  query    string  false  "搜索关键词 (昵称或手机号)"
// @Success      200  {object}  response.Response{data=types.GetUserListResp}  "用户列表"
// @Failure      200  {object}  response.Response{data=nil}                     "参数错误"
// @Router       /users [get]
func (h *UserHandler) GetUserList(c *gin.Context) {
	var req types.GetUserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errcode.ParamError, "分页参数或搜索关键词格式错误")
		return
	}

	total, users, err := h.svc.GetUserList(c.Request.Context(), req.Page, req.Size, req.Keyword)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	var list []types.UserProfileResp
	for _, u := range users {
		list = append(list, types.BuildUserProfileResp(u))
	}

	response.Success(
		c, types.GetUserListResp{
			Total: total,
			List:  list,
		},
	)
}

// GetUserInfo 获取指定用户信息
// @Summary      获取指定用户信息
// @Description  根据用户 UUID 获取其公开资料
// @Tags         用户
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "目标用户 UUID"
// @Success      200   {object}  response.Response{data=types.UserProfileResp}  "用户信息"
// @Failure      200   {object}  response.Response{data=nil}                     "用户不存在"
// @Router       /users/{uuid} [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	targetUUID := c.Param("uuid")
	if targetUUID == "" {
		response.Fail(c, errcode.ParamError, "用户ID不能为空")
		return
	}

	user, err := h.svc.GetUserByUUID(c.Request.Context(), targetUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, types.BuildUserProfileResp(user))
}

// SetUserStatus 启用/禁用用户 [管理员]
// @Summary      启用/禁用用户
// @Description  管理员操作：设置目标用户的状态为正常(0)或禁用(1)
// @Tags         用户管理 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string                true  "目标用户 UUID"
// @Param        req   body     types.SetUserStatusReq true  "状态值: 0=正常, 1=禁用"
// @Success      200   {object}  response.Response{data=nil}  "操作成功"
// @Failure      200   {object}  response.Response{data=nil}  "权限不足 / 参数错误"
// @Router       /users/{uuid}/status [put]
func (h *UserHandler) SetUserStatus(c *gin.Context) {
	if !middleware.CheckAdmin(c) {
		response.Fail(c, errcode.Unauthorized, "权限不足：需要管理员权限")
		return
	}

	targetUUID := c.Param("uuid")
	var req types.SetUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "状态值必须为 0(正常) 或 1(禁用)")
		return
	}

	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	err := h.svc.SetUserStatus(c.Request.Context(), operatorID, targetUUID, *req.Status)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// SetUserRole 设置用户角色 [管理员]
// @Summary      设置用户角色
// @Description  管理员操作：设置目标用户为普通用户(0)或管理员(1)
// @Tags         用户管理 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string               true  "目标用户 UUID"
// @Param        req   body     types.SetUserRoleReq  true  "角色值: 0=普通用户, 1=管理员"
// @Success      200   {object}  response.Response{data=nil}  "操作成功"
// @Failure      200   {object}  response.Response{data=nil}  "权限不足 / 参数错误"
// @Router       /users/{uuid}/role [put]
func (h *UserHandler) SetUserRole(c *gin.Context) {
	if !middleware.CheckAdmin(c) {
		response.Fail(c, errcode.Unauthorized, "权限不足：需要管理员权限")
		return
	}

	targetUUID := c.Param("uuid")
	var req types.SetUserRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "角色值必须为 0(普通用户) 或 1(管理员)")
		return
	}

	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	err := h.svc.SetUserRole(c.Request.Context(), operatorID, targetUUID, *req.IsAdmin)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteUser 删除用户 [管理员]
// @Summary      删除用户
// @Description  管理员操作：删除指定用户（不能删除自己）
// @Tags         用户管理 (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "目标用户 UUID"
// @Success      200   {object}  response.Response{data=nil}  "删除成功"
// @Failure      200   {object}  response.Response{data=nil}  "权限不足 / 不能删除自己"
// @Router       /users/{uuid} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	if !middleware.CheckAdmin(c) {
		response.Fail(c, errcode.Unauthorized, "权限不足：需要管理员权限")
		return
	}

	targetUUID := c.Param("uuid")
	if targetUUID == "" {
		response.Fail(c, errcode.ParamError, "用户ID不能为空")
		return
	}

	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	err := h.svc.DeleteUser(c.Request.Context(), operatorID, targetUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}
