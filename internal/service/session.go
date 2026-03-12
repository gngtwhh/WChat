package service

import (
    "context"
    "time"

    "github.com/rs/xid"

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
    groupRepo   *repository.GroupRepo
}

func NewSessionService(
    sessionRepo *repository.SessionRepo, userRepo *repository.UserRepo, groupRepo *repository.GroupRepo,
) *SessionService {
    return &SessionService{
        sessionRepo: sessionRepo,
        userRepo:    userRepo,
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

func (s *SessionService) CreateSession(ctx context.Context, userID, targetID string, sessionType int8) (string, error) {
    existSess, err := s.sessionRepo.FindSessionByTarget(ctx, userID, targetID, sessionType)
    if err != nil {
        return "", err
    }
    if existSess != nil {
        return existSess.Uuid, nil
    }

    now := time.Now()
    newSess := &model.Session{
        Uuid:          xid.New().String(),
        UserId:        userID,
        TargetId:      targetID,
        SessionType:   sessionType,
        LastMessageAt: &now,
    }

    if err := s.sessionRepo.Create(ctx, newSess); err != nil {
        return "", err
    }
    return newSess.Uuid, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, userID, sessionUUID string) error {
    sess, err := s.sessionRepo.FindByUUID(ctx, sessionUUID)
    if err != nil {
        return errcode.New(errcode.SessionNotFound)
    }
    if sess.UserId != userID {
        return errcode.New(errcode.Unauthorized)
    }
    return s.sessionRepo.DeleteByUUID(ctx, sessionUUID)
}

func (s *SessionService) SetSessionTopStatus(ctx context.Context, userID, sessionUUID string, isTop int8) error {
    sess, err := s.sessionRepo.FindByUUID(ctx, sessionUUID)
    if err != nil {
        return errcode.New(errcode.SessionNotFound)
    }
    if sess.UserId != userID {
        return errcode.New(errcode.Unauthorized)
    }
    return s.sessionRepo.UpdateField(ctx, sessionUUID, "is_top", isTop)
}

func (s *SessionService) ClearUnreadCount(ctx context.Context, userID, sessionUUID string) error {
    sess, err := s.sessionRepo.FindByUUID(ctx, sessionUUID)
    if err != nil {
        return errcode.New(errcode.SessionNotFound)
    }
    if sess.UserId != userID {
        return errcode.New(errcode.Unauthorized)
    }
    return s.sessionRepo.UpdateField(ctx, sessionUUID, "unread_count", 0)
}
