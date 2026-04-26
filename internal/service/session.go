package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"

	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"
)

type SessionInfo struct {
	Uuid          string
	TargetId      string
	TargetName    string
	TargetAvatar  string
	SessionType   int8
	UnreadCount   int
	LastMessage   string
	LastMessageAt string
	IsTop         int8
}

type SessionService struct {
	sessionRepo *repository.SessionRepo
	userRepo    *repository.UserRepo
	contactRepo *repository.ContactRepo
	groupRepo   *repository.GroupRepo
}

func NewSessionService(
	sessionRepo *repository.SessionRepo,
	userRepo *repository.UserRepo,
	contactRepo *repository.ContactRepo,
	groupRepo *repository.GroupRepo,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		contactRepo: contactRepo,
		groupRepo:   groupRepo,
	}
}

func (s *SessionService) GetSessionList(ctx context.Context, userID string, page, size int) (
	int64, []SessionInfo, error,
) {
	offset := (page - 1) * size
	total, sessions, err := s.sessionRepo.FindSessionList(ctx, userID, offset, size)
	if err != nil || total == 0 {
		return total, nil, err
	}

	var userIDs, groupIDs []string
	for _, sess := range sessions {
		if sess.SessionType == 0 {
			userIDs = append(userIDs, sess.TargetId)
		} else {
			groupIDs = append(groupIDs, sess.TargetId)
		}
	}

	users, _ := s.userRepo.FindByUUIDs(ctx, userIDs)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.Uuid] = u
	}

	groups, _ := s.groupRepo.FindByUUIDs(ctx, groupIDs)
	groupMap := make(map[string]model.Group)
	for _, g := range groups {
		groupMap[g.Uuid] = g
	}

	infos := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		info := SessionInfo{
			Uuid:        sess.Uuid,
			TargetId:    sess.TargetId,
			SessionType: sess.SessionType,
			UnreadCount: sess.UnreadCount,
			LastMessage: sess.LastMessage,
			IsTop:       sess.IsTop,
		}
		if sess.LastMessageAt != nil {
			info.LastMessageAt = sess.LastMessageAt.Format(time.RFC3339)
		}

		if sess.SessionType == 0 {
			info.TargetName = userMap[sess.TargetId].Nickname
			info.TargetAvatar = userMap[sess.TargetId].Avatar
		} else {
			info.TargetName = groupMap[sess.TargetId].Name
			info.TargetAvatar = groupMap[sess.TargetId].Avatar
		}
		infos = append(infos, info)
	}

	return total, infos, nil
}

func (s *SessionService) getOwnedSession(ctx context.Context, userID, sessionUUID string) (*model.Session, error) {
	sess, err := s.sessionRepo.FindActiveByUUID(ctx, sessionUUID)
	if err != nil {
		return nil, errcode.New(errcode.SessionNotFound)
	}
	if sess.UserId != userID {
		return nil, errcode.New(errcode.Unauthorized)
	}
	return sess, nil
}

func (s *SessionService) CreateSession(ctx context.Context, userID, targetID string, sessionType int8) (string, error) {
	switch sessionType {
	case 0:
		if targetID == userID {
			return "", errcode.New(errcode.CannotAddSelf)
		}
		if _, err := s.userRepo.FindActiveByUUID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", errcode.New(errcode.UserNotFound)
			}
			return "", err
		}
		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, userID, targetID)
		if err != nil {
			return "", err
		}
		if !isFriend {
			return "", errcode.New(errcode.ContactNotFound)
		}
	case 1:
		group, err := s.groupRepo.FindActiveByUUID(ctx, targetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", errcode.New(errcode.GroupNotFound)
			}
			return "", err
		}
		var members []string
		if err := json.Unmarshal(group.Members, &members); err != nil {
			return "", err
		}
		inGroup := false
		for _, memberUUID := range members {
			if memberUUID == userID {
				inGroup = true
				break
			}
		}
		if !inGroup {
			return "", errcode.New(errcode.NotInGroup)
		}
	default:
		return "", errcode.New(errcode.ParamError)
	}

	now := time.Now()
	sess, err := s.sessionRepo.FindOrCreate(ctx, userID, targetID, sessionType, xid.New().String(), &now)
	if err != nil {
		return "", err
	}
	return sess.Uuid, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, userID, sessionUUID string) error {
	if _, err := s.getOwnedSession(ctx, userID, sessionUUID); err != nil {
		return err
	}
	return s.sessionRepo.SoftDeleteByUUID(ctx, sessionUUID)
}

func (s *SessionService) SetSessionTopStatus(ctx context.Context, userID, sessionUUID string, isTop int8) error {
	if _, err := s.getOwnedSession(ctx, userID, sessionUUID); err != nil {
		return err
	}
	return s.sessionRepo.UpdateTopStatusByUUID(ctx, sessionUUID, isTop)
}

func (s *SessionService) ClearUnreadCount(ctx context.Context, userID, sessionUUID string) error {
	if _, err := s.getOwnedSession(ctx, userID, sessionUUID); err != nil {
		return err
	}
	return s.sessionRepo.ClearUnreadCountByUUID(ctx, sessionUUID)
}
