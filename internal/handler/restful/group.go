package restful

import (
	"github.com/gin-gonic/gin"

	"wchat/internal/middleware"
	"wchat/internal/service"
	"wchat/internal/types"
	"wchat/pkg/errcode"
	"wchat/pkg/response"
)

type GroupHandler struct {
	svc *service.GroupService
}

func NewGroupHandler(svc *service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// CreateGroup 创建群聊
// @Summary      创建群聊
// @Description  创建一个新的群聊，当前用户自动成为群主；可选择在创建时邀请初始成员
// @Tags         群组
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req  body      types.CreateGroupReq                      true  "群组信息"
// @Success      200  {object}  response.Response{data=types.CreateGroupResp}  "返回群 UUID"
// @Failure      200  {object}  response.Response{data=nil}                     "参数错误"
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	ownerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	var req types.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "群组参数格式错误")
		return
	}

	groupUUID, err := h.svc.CreateGroup(
		c.Request.Context(), ownerID, req.Name, req.Notice, req.Avatar, req.AddMode, req.Members,
	)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, types.CreateGroupResp{Uuid: groupUUID})
}

// GetJoinedGroups 获取我加入的群列表
// @Summary      获取我加入的群列表
// @Description  获取当前用户已加入的所有群聊，包含群名、成员数、群主等信息
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=types.GetJoinedGroupsResp}  "群组列表"
// @Failure      200  {object}  response.Response{data=nil}                         "Token 无效"
// @Router       /groups/joined [get]
func (h *GroupHandler) GetJoinedGroups(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupInfos, err := h.svc.GetJoinedGroups(c.Request.Context(), userID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	voList := make([]types.GroupVO, 0, len(groupInfos))
	for _, info := range groupInfos {
		voList = append(
			voList, types.GroupVO{
				Uuid:      info.Uuid,
				Name:      info.Name,
				Notice:    info.Notice,
				Avatar:    info.Avatar,
				MemberCnt: info.MemberCnt,
				OwnerId:   info.OwnerId,
				AddMode:   info.AddMode,
				Status:    info.Status,
			},
		)
	}

	response.Success(
		c, types.GetJoinedGroupsResp{
			Total: len(voList),
			List:  voList,
		},
	)
}

// GetGroupInfo 获取群详情
// @Summary      获取群详情
// @Description  根据群 UUID 获取群名称、公告、头像、成员数、群主、加群模式等信息
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "群 UUID"
// @Success      200   {object}  response.Response{data=types.GroupVO}  "群详情"
// @Failure      200   {object}  response.Response{data=nil}            "群不存在"
// @Router       /groups/{uuid} [get]
func (h *GroupHandler) GetGroupInfo(c *gin.Context) {
	groupUUID := c.Param("uuid")
	if groupUUID == "" {
		response.Fail(c, errcode.ParamError, "群组ID不能为空")
		return
	}

	info, err := h.svc.GetGroupInfo(c.Request.Context(), groupUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(
		c, types.GroupVO{
			Uuid:      info.Uuid,
			Name:      info.Name,
			Notice:    info.Notice,
			Avatar:    info.Avatar,
			MemberCnt: info.MemberCnt,
			OwnerId:   info.OwnerId,
			AddMode:   info.AddMode,
			Status:    info.Status,
		},
	)
}

// UpdateGroupInfo 修改群资料
// @Summary      修改群资料
// @Description  群主修改群名称、公告、头像、加群模式等（仅传入需修改的字段）
// @Tags         群组
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string                true  "群 UUID"
// @Param        req   body     types.UpdateGroupReq   true  "需要更新的字段"
// @Success      200   {object}  response.Response{data=nil}  "修改成功"
// @Failure      200   {object}  response.Response{data=nil}  "权限不足 / 群不存在"
// @Router       /groups/{uuid} [put]
func (h *GroupHandler) UpdateGroupInfo(c *gin.Context) {
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupUUID := c.Param("uuid")

	var req types.UpdateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "参数格式错误")
		return
	}

	dto := service.GroupUpdateDTO{
		Name:    req.Name,
		Notice:  req.Notice,
		Avatar:  req.Avatar,
		AddMode: req.AddMode,
	}

	err := h.svc.UpdateGroupInfo(c.Request.Context(), operatorID, groupUUID, dto)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// DismissGroup 解散群聊
// @Summary      解散群聊
// @Description  群主解散群聊（仅群主有权限），解散后群内所有会话失效
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "群 UUID"
// @Success      200   {object}  response.Response{data=nil}  "解散成功"
// @Failure      200   {object}  response.Response{data=nil}  "权限不足 / 群不存在"
// @Router       /groups/{uuid} [delete]
func (h *GroupHandler) DismissGroup(c *gin.Context) {
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupUUID := c.Param("uuid")

	err := h.svc.DismissGroup(c.Request.Context(), operatorID, groupUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// GetGroupMembers 获取群成员列表
// @Summary      获取群成员列表
// @Description  获取指定群聊的所有成员，包含昵称、头像、签名等信息
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "群 UUID"
// @Success      200   {object}  response.Response{data=types.GetGroupMembersResp}  "成员列表"
// @Failure      200   {object}  response.Response{data=nil}                         "群不存在"
// @Router       /groups/{uuid}/members [get]
func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupUUID := c.Param("uuid")

	members, err := h.svc.GetGroupMembers(c.Request.Context(), groupUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	voList := make([]types.GroupMemberVO, 0, len(members))
	for _, m := range members {
		voList = append(
			voList, types.GroupMemberVO{
				Uuid:      m.Uuid,
				Nickname:  m.Nickname,
				Avatar:    m.Avatar,
				Signature: m.Signature,
			},
		)
	}

	response.Success(
		c, types.GetGroupMembersResp{
			Total: len(voList),
			List:  voList,
		},
	)
}

// InviteToGroup 邀请成员入群
// @Summary      邀请成员入群
// @Description  邀请一个或多个用户加入指定群聊
// @Tags         群组
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string                 true  "群 UUID"
// @Param        req   body     types.InviteMemberReq   true  "被邀请的用户 UUID 列表"
// @Success      200   {object}  response.Response{data=nil}  "邀请成功"
// @Failure      200   {object}  response.Response{data=nil}  "参数错误 / 群不存在"
// @Router       /groups/{uuid}/members [post]
func (h *GroupHandler) InviteToGroup(c *gin.Context) {
	inviterID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupUUID := c.Param("uuid")
	var req types.InviteMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ParamError, "请提供要邀请的用户列表")
		return
	}

	err := h.svc.InviteToGroup(c.Request.Context(), inviterID, groupUUID, req.UserIds)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// LeaveGroup 退出群聊
// @Summary      退出群聊
// @Description  当前用户主动退出指定群聊（群主不能退出，需先转让或解散）
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path     string  true  "群 UUID"
// @Success      200   {object}  response.Response{data=nil}  "退出成功"
// @Failure      200   {object}  response.Response{data=nil}  "不在群中 / 群主不能退出"
// @Router       /groups/{uuid}/members/me [delete]
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupUUID := c.Param("uuid")

	err := h.svc.LeaveGroup(c.Request.Context(), userID, groupUUID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}

// KickMember 踢出群成员
// @Summary      踢出群成员
// @Description  群主/管理员将指定用户从群聊中移除（不能踢自己）
// @Tags         群组
// @Produce      json
// @Security     BearerAuth
// @Param        uuid     path     string  true  "群 UUID"
// @Param        user_id  path     string  true  "被踢用户 UUID"
// @Success      200      {object}  response.Response{data=nil}  "踢出成功"
// @Failure      200      {object}  response.Response{data=nil}  "权限不足 / 不能踢自己"
// @Router       /groups/{uuid}/members/{user_id} [delete]
func (h *GroupHandler) KickMember(c *gin.Context) {
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	groupUUID := c.Param("uuid")
	targetUserID := c.Param("user_id")

	if operatorID == targetUserID {
		response.Fail(c, errcode.ParamError, "不能踢出自己，请使用退群功能")
		return
	}

	err := h.svc.KickMember(c.Request.Context(), operatorID, groupUUID, targetUserID)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	response.Success(c, nil)
}
