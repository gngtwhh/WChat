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
		response.Fail(c, errcode.ParamError, "target user id cannot be empty")
		return
	}

	if userID == targetUUID {
		response.Fail(c, errcode.ParamError, "cannot operate on yourself")
		return
	}

	if err := h.svc.DeleteContact(c.Request.Context(), userID, targetUUID); err != nil {
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
		response.Fail(c, errcode.ParamError, "target user id cannot be empty")
		return
	}

	if userID == targetUUID {
		response.Fail(c, errcode.ParamError, "cannot operate on yourself")
		return
	}

	if err := h.svc.BlockContact(c.Request.Context(), userID, targetUUID); err != nil {
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
		response.Fail(c, errcode.ParamError, "target user id cannot be empty")
		return
	}

	if err := h.svc.UnblockContact(c.Request.Context(), userID, targetUUID); err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}
