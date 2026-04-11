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

// GetMessageList 获取消息列表
// @Summary      获取消息列表
// @Description  分页获取指定会话的历史消息，按发送时间倒序返回
// @Tags         消息
// @Produce      json
// @Security     BearerAuth
// @Param        session_id  query    string  true  "会话 UUID"
// @Param        page        query    int     true  "页码 (从1开始)"  minimum(1)
// @Param        size        query    int     true  "每页数量"        minimum(1) maximum(100)
// @Success      200  {object}  response.Response{data=types.GetMessageListResp}  "消息列表"
// @Failure      200  {object}  response.Response{data=nil}                        "参数错误 / Token 无效"
// @Router       /messages [get]
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

// RecallMessage 撤回消息
// @Summary      撤回消息
// @Description  撤回自己发送的消息（将消息状态设为 2=已撤回），超时可能被拒绝
// @Tags         消息
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "消息 UUID"
// @Success      200   {object}  response.Response{data=nil}  "撤回成功"
// @Failure      200   {object}  response.Response{data=nil}  "消息不存在 / 超时 / 非本人消息"
// @Router       /messages/{uuid}/recall [put]
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
