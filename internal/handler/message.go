package handler

import (
    "github.com/gin-gonic/gin"

    "wchat/internal/middleware"
    "wchat/internal/service"
    "wchat/internal/types"
    "wchat/pkg/errcode"
    "wchat/pkg/response"
)

type MessageHandler struct {
    svc *service.MessageService
}

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
    return &MessageHandler{svc: svc}
}

func (h *MessageHandler) GetMessageList(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    var req types.GetMessageListReq
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Fail(c, errcode.ParamError, "请提供正确的 session_id 和分页参数")
        return
    }

    total, infos, err := h.svc.GetMessageList(c.Request.Context(), userID, req.SessionId, req.Page, req.Size)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    voList := make([]types.MessageVO, 0, len(infos))
    for _, info := range infos {
        voList = append(
            voList, types.MessageVO{
                Uuid:      info.Uuid,
                SessionId: info.SessionId,
                Type:      info.Type,
                Content:   info.Content,
                Url:       info.Url,
                SendId:    info.SendId,
                ReceiveId: info.ReceiveId,
                FileType:  info.FileType,
                FileName:  info.FileName,
                FileSize:  info.FileSize,
                AVdata:    info.AVdata,
                Status:    info.Status,
                SendAt:    info.SendAt,
            },
        )
    }

    response.Success(
        c, types.GetMessageListResp{
            Total: total,
            List:  voList,
        },
    )
}

func (h *MessageHandler) RecallMessage(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    msgUUID := c.Param("uuid")
    if msgUUID == "" {
        response.Fail(c, errcode.ParamError, "消息ID不能为空")
        return
    }

    if err := h.svc.RecallMessage(c.Request.Context(), userID, msgUUID); err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}
