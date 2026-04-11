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

// GetSessionList 获取会话列表
// @Summary      获取会话列表
// @Description  分页获取当前用户的聊天会话列表，包含对方名称、头像、最后消息、未读数等
// @Tags         会话
// @Produce      json
// @Security     BearerAuth
// @Param        page  query    int  true  "页码 (从1开始)"  minimum(1)
// @Param        size  query    int  true  "每页数量"        minimum(1) maximum(100)
// @Success      200   {object}  response.Response{data=types.GetSessionListResp}  "会话列表"
// @Failure      200   {object}  response.Response{data=nil}                        "参数错误 / Token 无效"
// @Router       /sessions [get]
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

// CreateSession 创建会话
// @Summary      创建会话
// @Description  与目标用户或群组创建一个新会话（如果已存在则返回已有会话的 UUID）
// @Tags         会话
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.CreateSessionReq                    true  "target_id=对方UUID, session_type: 0=私聊, 1=群聊"
// @Success      200  {object}  response.Response{data=types.CreateSessionResp}  "返回会话 UUID"
// @Failure      200  {object}  response.Response{data=nil}                       "不能与自己发起会话 / 参数错误"
// @Router       /sessions [post]
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

	sessionUUID, err := h.svc.CreateSession(c.Request.Context(), userID, req.TargetId, req.SessionType)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, types.CreateSessionResp{Uuid: sessionUUID})
}

// DeleteSession 删除会话
// @Summary      删除会话
// @Description  从当前用户的会话列表中删除指定会话
// @Tags         会话
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "会话 UUID"
// @Success      200   {object}  response.Response{data=nil}  "删除成功"
// @Failure      200   {object}  response.Response{data=nil}  "会话不存在"
// @Router       /sessions/{uuid} [delete]
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

// PinSession 置顶/取消置顶会话
// @Summary      置顶/取消置顶会话
// @Description  设置指定会话的置顶状态：1=置顶, 0=取消置顶
// @Tags         会话
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string               true  "会话 UUID"
// @Param        req   body     types.PinSessionReq   true  "is_top: 0=取消置顶, 1=置顶"
// @Success      200   {object}  response.Response{data=nil}  "操作成功"
// @Failure      200   {object}  response.Response{data=nil}  "参数错误"
// @Router       /sessions/{uuid}/top [put]
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

// ClearUnreadCount 标记会话已读
// @Summary      标记会话已读
// @Description  将指定会话的未读消息数归零
// @Tags         会话
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "会话 UUID"
// @Success      200   {object}  response.Response{data=nil}  "标记成功"
// @Failure      200   {object}  response.Response{data=nil}  "会话不存在"
// @Router       /sessions/{uuid}/read [put]
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
