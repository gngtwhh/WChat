package handler

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
