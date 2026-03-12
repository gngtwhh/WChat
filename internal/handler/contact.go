package handler

import (
    "github.com/gin-gonic/gin"

    "wchat/internal/middleware"
    "wchat/internal/service"
    "wchat/internal/types"
    "wchat/pkg/errcode"
    "wchat/pkg/response"
)

type ContactHandler struct {
    svc *service.ContactService
}

func NewContactHandler(svc *service.ContactService) *ContactHandler {
    return &ContactHandler{svc: svc}
}

func (h *ContactHandler) GetContactList(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    // 1. 获取 Service 层的纯领域数据
    contactInfos, err := h.svc.GetUserContactList(c.Request.Context(), userID)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    voList := make([]types.ContactVO, 0, len(contactInfos))
    for _, info := range contactInfos {
        voList = append(
            voList, types.ContactVO{
                ContactId: info.ContactId,
                Nickname:  info.Nickname,
                Avatar:    info.Avatar,
                Signature: info.Signature,
                Status:    info.Status,
            },
        )
    }

    response.Success(
        c, types.GetContactListResp{
            Total: len(voList),
            List:  voList,
        },
    )
}

func (h *ContactHandler) DeleteContact(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    targetUUID := c.Param("uuid")
    if targetUUID == "" {
        response.Fail(c, errcode.ParamError, "目标用户ID不能为空")
        return
    }

    if userID == targetUUID {
        response.Fail(c, errcode.ParamError, "不能对自身进行操作")
        return
    }

    // 状态 3: 删除好友，状态 4: 被删除好友
    err := h.svc.UpdateBiDirectionalStatus(c.Request.Context(), userID, targetUUID, 3, 4)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}

func (h *ContactHandler) BlockContact(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    targetUUID := c.Param("uuid")
    if targetUUID == "" {
        response.Fail(c, errcode.ParamError, "目标用户ID不能为空")
        return
    }

    if userID == targetUUID {
        response.Fail(c, errcode.ParamError, "不能对自己执行拉黑操作")
        return
    }

    // 状态 1: 拉黑，状态 2: 被拉黑
    err := h.svc.UpdateBiDirectionalStatus(c.Request.Context(), userID, targetUUID, 1, 2)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}

func (h *ContactHandler) UnblockContact(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    targetUUID := c.Param("uuid")
    if targetUUID == "" {
        response.Fail(c, errcode.ParamError, "目标用户ID不能为空")
        return
    }

    // 恢复正常状态 (双向恢复为 0)
    err := h.svc.UpdateBiDirectionalStatus(c.Request.Context(), userID, targetUUID, 0, 0)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}
