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

// ==============================================================
// 6. 设置管理员 (SetUserRole) [Admin]
// ==============================================================

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

// ==============================================================
// 7. 删除用户 (DeleteUser) [Admin]
// ==============================================================

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
