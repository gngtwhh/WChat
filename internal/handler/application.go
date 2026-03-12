package handler

import (
    "github.com/gin-gonic/gin"

    "wchat/internal/middleware"
    "wchat/internal/service"
    "wchat/internal/types"
    "wchat/pkg/errcode"
    "wchat/pkg/response"
)

type ApplicationHandler struct {
    svc *service.ApplicationService
}

func NewApplicationHandler(svc *service.ApplicationService) *ApplicationHandler {
    return &ApplicationHandler{svc: svc}
}

func (h *ApplicationHandler) GetApplicationList(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    applies, err := h.svc.GetApplicationList(c.Request.Context(), userID)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    voList := make([]types.ApplicationVO, 0, len(applies))
    for _, app := range applies {
        voList = append(
            voList, types.ApplicationVO{
                Uuid:        app.Uuid,
                UserId:      app.UserId,
                Nickname:    app.Nickname,
                Avatar:      app.Avatar,
                ContactId:   app.ContactId,
                ContactType: app.ContactType,
                Status:      app.Status,
                Message:     app.Message,
                LastApplyAt: app.LastApplyAt,
            },
        )
    }

    response.Success(
        c, types.GetApplicationListResp{
            Total: len(voList),
            List:  voList,
        },
    )
}

// SubmitApplication 发起加好友/加群申请
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
    applicantID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    var req types.SubmitApplicationReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "参数格式错误")
        return
    }

    err := h.svc.SubmitApplication(c.Request.Context(), applicantID, req.ContactId, req.ContactType, req.Message)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}

func (h *ApplicationHandler) HandleApplication(c *gin.Context) {
    operatorID, ok := middleware.GetUserID(c)
    if !ok {
        response.Fail(c, errcode.TokenInvalid)
        return
    }

    applyUUID := c.Param("uuid")
    if applyUUID == "" {
        response.Fail(c, errcode.ParamError, "申请ID不能为空")
        return
    }

    var req types.HandleApplicationReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, errcode.ParamError, "状态参数不合法")
        return
    }

    err := h.svc.HandleApplication(c.Request.Context(), operatorID, applyUUID, req.Status)
    if err != nil {
        response.FailErr(c, err)
        return
    }

    response.Success(c, nil)
}
