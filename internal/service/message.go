package service

import (
    "context"
    "time"

    "wchat/internal/repository"
    "wchat/pkg/errcode"
)

type MessageInfo struct {
    Uuid      string
    SessionId string
    Type      int8
    Content   string
    Url       string
    SendId    string
    ReceiveId string
    FileType  string
    FileName  string
    FileSize  int64
    AVdata    string
    Status    int8
    SendAt    string
}

type MessageService struct {
    messageRepo *repository.MessageRepo
    sessionRepo *repository.SessionRepo
}

func NewMessageService(messageRepo *repository.MessageRepo, sessionRepo *repository.SessionRepo) *MessageService {
    return &MessageService{
        messageRepo: messageRepo,
        sessionRepo: sessionRepo,
    }
}

func (s *MessageService) GetMessageList(ctx context.Context, userID, sessionID string, page, size int) (
    int64, []MessageInfo, error,
) {
    sess, err := s.sessionRepo.FindByUUID(ctx, sessionID)
    if err != nil {
        return 0, nil, errcode.New(errcode.SessionNotFound)
    }
    if sess.UserId != userID {
        return 0, nil, errcode.New(errcode.Unauthorized)
    }

    offset := (page - 1) * size
    total, messages, err := s.messageRepo.FindMessageList(ctx, sessionID, offset, size)
    if err != nil || total == 0 {
        return total, nil, err
    }

    infos := make([]MessageInfo, 0, len(messages))
    for _, msg := range messages {
        info := MessageInfo{
            Uuid:      msg.Uuid,
            SessionId: msg.SessionId,
            Type:      msg.Type,
            Content:   msg.Content,
            Url:       msg.Url,
            SendId:    msg.SendId,
            ReceiveId: msg.ReceiveId,
            FileType:  msg.FileType,
            FileName:  msg.FileName,
            FileSize:  msg.FileSize,
            AVdata:    msg.AVdata,
            Status:    msg.Status,
        }
        if msg.SendAt != nil {
            info.SendAt = msg.SendAt.Format(time.RFC3339)
        }
        infos = append(infos, info)
    }

    return total, infos, nil
}

func (s *MessageService) RecallMessage(ctx context.Context, userID, msgUUID string) error {
    msg, err := s.messageRepo.FindByUUID(ctx, msgUUID)
    if err != nil {
        return errcode.New(errcode.MessageNotFound)
    }

    if msg.SendId != userID {
        return errcode.New(errcode.Unauthorized)
    }

    if msg.SendAt != nil {
        if time.Since(*msg.SendAt) > 2*time.Minute {
            return errcode.New(errcode.MessageRecallTimeout)
        }
    }

    // 状态 2: 已撤回
    return s.messageRepo.UpdateStatus(ctx, msgUUID, 2)
}
