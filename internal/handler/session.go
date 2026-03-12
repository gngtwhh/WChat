package handler

import (
    "github.com/gin-gonic/gin"

    "wchat/internal/middleware"
    "wchat/internal/service"
    "wchat/internal/types"
    "wchat/pkg/errcode"
    "wchat/pkg/response"
)

type SessionHandler struct {
    svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
    return &SessionHandler{svc: svc}
}

func (h *SessionHandler) GetSessionList(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    var req types.GetSessionListReq
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Fail(c, errcode.ParamError, "分页参数格式错误")
        return
    }

    total, infos, err := h.svc.GetSessionList(c.Request.Context(), userID, req.Page, req.Size)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    voList := make([]types.SessionVO, 0, len(infos))
    for _, info := range infos {
        voList = append(
            voList, types.SessionVO{
                Uuid:          info.Uuid,
                TargetId:      info.TargetId,
                TargetName:    info.TargetName,
                TargetAvatar:  info.TargetAvatar,
                SessionType:   info.SessionType,
                UnreadCount:   info.UnreadCount,
                LastMessage:   info.LastMessage,
                LastMessageAt: info.LastMessageAt,
                IsTop:         info.IsTop,
            },
        )
    }

    response.Success(
        c, types.GetSessionListResp{
            Total: total,
            List:  voList,
        },
    )
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    var req types.CreateSessionReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "目标ID或会话类型错误")
        return
    }

    if req.SessionType == 0 && userID == req.TargetId {
        response.Fail(c, errcode.ParamError, "不能与自己发起会话")
        return
    }

    sessionUUID, err := h.svc.CreateSession(c.Request.Context(), userID, req.TargetId, req.SessionType)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, types.CreateSessionResp{Uuid: sessionUUID})
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    sessionUUID := c.Param("uuid")
    if sessionUUID == "" {
        response.Fail(c, errcode.ParamError, "会话ID不能为空")
        return
    }

    if err := h.svc.DeleteSession(c.Request.Context(), userID, sessionUUID); err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}

func (h *SessionHandler) PinSession(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    sessionUUID := c.Param("uuid")

    var req types.PinSessionReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "状态值必须为 0 或 1")
        return
    }

    if err := h.svc.SetSessionTopStatus(c.Request.Context(), userID, sessionUUID, *req.IsTop); err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}

func (h *SessionHandler) ClearUnreadCount(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    sessionUUID := c.Param("uuid")
    if sessionUUID == "" {
        response.Fail(c, errcode.ParamError, "会话ID不能为空")
        return
    }

    if err := h.svc.ClearUnreadCount(c.Request.Context(), userID, sessionUUID); err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}
