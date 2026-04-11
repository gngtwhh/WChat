package service

import (
	"context"
	"encoding/json"

	"github.com/rs/xid"

	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"
)

type GroupInfo struct {
	Uuid      string
	Name      string
	Notice    string
	Avatar    string
	MemberCnt int
	OwnerId   string
	AddMode   int8
	Status    int8
}

type GroupMemberInfo struct {
	Uuid      string
	Nickname  string
	Avatar    string
	Signature string
}

type GroupUpdateDTO struct {
	Name    string
	Notice  string
	Avatar  string
	AddMode *int8
}

type GroupService struct {
	groupRepo   *repository.GroupRepo
	userRepo    *repository.UserRepo
	contactRepo *repository.ContactRepo
}

func NewGroupService(
	groupRepo *repository.GroupRepo, userRepo *repository.UserRepo, contactRepo *repository.ContactRepo,
) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		contactRepo: contactRepo,
	}
}

func (s *GroupService) getValidGroup(ctx context.Context, groupUUID string) (*model.Group, error) {
	group, err := s.groupRepo.FindActiveByUUID(ctx, groupUUID)
	if err != nil {
		return nil, errcode.New(errcode.GroupNotFound)
	}
	if group.Status == 2 {
		return nil, errcode.New(errcode.GroupDismissed)
	}
	return group, nil
}

func (s *GroupService) CreateGroup(
	ctx context.Context, ownerID string, name, notice, avatar string, addMode *int8, initialMembers []string,
) (string, error) {
	mode := int8(0)
	if addMode != nil {
		mode = *addMode
	}

	groupUUID := xid.New().String()
	memberMap := map[string]bool{ownerID: true}
	var finalMembers = []string{ownerID}

	for _, m := range initialMembers {
		if m == ownerID {
			continue
		}

		// 必须是好友
		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, ownerID, m)
		if err != nil {
			return "", err
		}
		if !isFriend {
			return "", errcode.New(errcode.ContactNotFound) // 你们还不是好友
		}

		if !memberMap[m] {
			finalMembers = append(finalMembers, m)
			memberMap[m] = true
		}
	}

	// TODO: 如果后期有群组最大人数限制，可以在这里校验并返回 errcode.GroupFull

	membersJSON, _ := json.Marshal(finalMembers)
	group := &model.Group{
		Uuid:      groupUUID,
		Name:      name,
		Notice:    notice,
		Members:   membersJSON,
		MemberCnt: len(finalMembers),
		OwnerId:   ownerID,
		AddMode:   mode,
		Avatar:    avatar,
		Status:    0,
	}

	if err := s.groupRepo.CreateWithMembers(ctx, group, finalMembers); err != nil {
		return "", err
	}

	return groupUUID, nil
}

func (s *GroupService) GetJoinedGroups(ctx context.Context, userID string) ([]GroupInfo, error) {
	groups, err := s.groupRepo.FindJoinedGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	infos := make([]GroupInfo, 0, len(groups))
	for _, g := range groups {
		infos = append(
			infos, GroupInfo{
				Uuid: g.Uuid, Name: g.Name, Notice: g.Notice, Avatar: g.Avatar,
				MemberCnt: g.MemberCnt, OwnerId: g.OwnerId, AddMode: g.AddMode, Status: g.Status,
			},
		)
	}
	return infos, nil
}

func (s *GroupService) GetGroupInfo(ctx context.Context, groupUUID string) (GroupInfo, error) {
	g, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return GroupInfo{}, err
	}
	return GroupInfo{
		Uuid: g.Uuid, Name: g.Name, Notice: g.Notice, Avatar: g.Avatar,
		MemberCnt: g.MemberCnt, OwnerId: g.OwnerId, AddMode: g.AddMode, Status: g.Status,
	}, nil
}

func (s *GroupService) UpdateGroupInfo(ctx context.Context, operatorID, groupUUID string, dto GroupUpdateDTO) error {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return err
	}

	if group.OwnerId != operatorID {
		return errcode.New(errcode.NoGroupPermission) // 只有群主或管理员才能执行此操作
	}

	updateData := make(map[string]any)
	if dto.Name != "" {
		updateData["name"] = dto.Name
	}
	if dto.Notice != "" {
		updateData["notice"] = dto.Notice
	}
	if dto.Avatar != "" {
		updateData["avatar"] = dto.Avatar
	}
	if dto.AddMode != nil {
		updateData["add_mode"] = *dto.AddMode
	}

	if len(updateData) == 0 {
		return nil
	}

	return s.groupRepo.UpdateInfoByUUID(ctx, groupUUID, updateData)
}

// TODO: 目前为修改群状态，彻底删除该群聊记录的操作应该在何时处理？
func (s *GroupService) DismissGroup(ctx context.Context, operatorID, groupUUID string) error {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return err
	}
	if group.OwnerId != operatorID {
		return errcode.New(errcode.NoGroupPermission) // 只有群主或管理员才能执行此操作
	}

	return s.groupRepo.DismissByUUID(ctx, groupUUID)
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupUUID string) ([]GroupMemberInfo, error) {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	var memberUUIDs []string
	if err := json.Unmarshal(group.Members, &memberUUIDs); err != nil {
		return nil, err
	}

	users, err := s.userRepo.FindByUUIDs(ctx, memberUUIDs)
	if err != nil {
		return nil, err
	}

	infos := make([]GroupMemberInfo, 0, len(users))
	for _, u := range users {
		infos = append(
			infos, GroupMemberInfo{
				Uuid: u.Uuid, Nickname: u.Nickname, Avatar: u.Avatar, Signature: u.Signature,
			},
		)
	}
	return infos, nil
}

func (s *GroupService) InviteToGroup(ctx context.Context, inviterID, groupUUID string, userIDs []string) error {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return err
	}

	// 获取当前已经在群里的成员列表，用于排重
	var currentMembers []string
	_ = json.Unmarshal(group.Members, &currentMembers)
	inGroupMap := make(map[string]bool)
	for _, m := range currentMembers {
		inGroupMap[m] = true
	}

	for _, uid := range userIDs {
		if uid == inviterID {
			return errcode.New(errcode.CannotAddSelf) // 不能添加自己
		}
		if inGroupMap[uid] {
			return errcode.New(errcode.AlreadyInGroup) // 你(ta)已经在此群聊中
		}

		// 必须是好友
		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, inviterID, uid)
		if err != nil {
			return err
		}
		if !isFriend {
			return errcode.New(errcode.ContactNotFound) // 你们还不是好友
		}
	}

	if group.AddMode == 1 && group.OwnerId != inviterID {
		return errcode.NewWithMsg(errcode.NoGroupPermission, "该群已开启审核模式，请等待群主审批功能上线")
	}

	return s.groupRepo.AddMembers(ctx, groupUUID, userIDs)
}

func (s *GroupService) LeaveGroup(ctx context.Context, userID, groupUUID string) error {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return err
	}

	if group.OwnerId == userID {
		return errcode.NewWithMsg(errcode.NoGroupPermission, "群主无法直接退群，请先解散或转让群组")
	}

	// 校验是否在群内
	var currentMembers []string
	_ = json.Unmarshal(group.Members, &currentMembers)
	isInGroup := false
	for _, m := range currentMembers {
		if m == userID {
			isInGroup = true
			break
		}
	}
	if !isInGroup {
		return errcode.New(errcode.NotInGroup) // 你不在该群聊中
	}

	return s.groupRepo.RemoveMember(ctx, groupUUID, userID, 6)
}

func (s *GroupService) KickMember(ctx context.Context, operatorID, groupUUID, targetUserID string) error {
	group, err := s.getValidGroup(ctx, groupUUID)
	if err != nil {
		return err
	}

	if group.OwnerId != operatorID {
		return errcode.New(errcode.NoGroupPermission) // 只有群主或管理员才能执行此操作
	}

	// 校验目标是否在群内
	var currentMembers []string
	_ = json.Unmarshal(group.Members, &currentMembers)
	isInGroup := false
	for _, m := range currentMembers {
		if m == targetUserID {
			isInGroup = true
			break
		}
	}
	if !isInGroup {
		return errcode.NewWithMsg(errcode.NotInGroup, "目标用户不在该群聊中")
	}

	return s.groupRepo.RemoveMember(ctx, groupUUID, targetUserID, 7)
}
