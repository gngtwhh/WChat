package service

import (
	"context"
	"encoding/json"
	"time"

	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"

	"github.com/rs/xid"
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

type SendMessageResult struct {
	Message             *MessageInfo
	RecipientIDs        []string
	RecipientSessionIDs map[string]string
}

type MessageService struct {
	messageRepo *repository.MessageRepo
	sessionRepo *repository.SessionRepo
	contactRepo *repository.ContactRepo
	groupRepo   *repository.GroupRepo
}

func NewMessageService(
	messageRepo *repository.MessageRepo,
	sessionRepo *repository.SessionRepo,
	contactRepo *repository.ContactRepo,
	groupRepo *repository.GroupRepo,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		sessionRepo: sessionRepo,
		contactRepo: contactRepo,
		groupRepo:   groupRepo,
	}
}

func (s *MessageService) GetMessageList(ctx context.Context, userID, sessionID string, page, size int) (
	int64, []MessageInfo, error,
) {
	sess, err := s.sessionRepo.FindActiveByUUID(ctx, sessionID)
	if err != nil {
		return 0, nil, errcode.New(errcode.SessionNotFound)
	}
	if sess.UserId != userID {
		return 0, nil, errcode.New(errcode.Unauthorized)
	}

	offset := (page - 1) * size
	conversationID := buildConversationID(sess.SessionType, sess.UserId, sess.TargetId)
	total, messages, err := s.messageRepo.FindMessageListByConversation(ctx, conversationID, offset, size)
	if err != nil || total == 0 {
		return total, nil, err
	}

	infos := make([]MessageInfo, 0, len(messages))
	for _, msg := range messages {
		info := MessageInfo{
			Uuid:      msg.Uuid,
			SessionId: sessionID,
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
	msg, err := s.messageRepo.FindActiveByUUID(ctx, msgUUID)
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
	return s.messageRepo.MarkRecalledByUUID(ctx, msgUUID)
}

func (s *MessageService) ensureSession(
	ctx context.Context, userID, targetID string, sessionType int8, now time.Time,
) (*model.Session, error) {
	return s.sessionRepo.FindOrCreateByTarget(ctx, userID, targetID, sessionType, xid.New().String(), &now)
}

// SendMessage handles message persistence, validation, and session updates.
func (s *MessageService) SendMessage(
	ctx context.Context, sendID, receiveID string, sessionType, msgType int8, content, url string,
) (*SendMessageResult, error) {
	// Permission check
	recipientIDs := make([]string, 0, 1)
	if sessionType == 0 {
		// Private chat: check if they are contacts
		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, sendID, receiveID)
		if err != nil || !isFriend {
			return nil, errcode.New(errcode.ContactNotFound)
		}
		recipientIDs = append(recipientIDs, receiveID)
	} else {
		// Group chat: check if sender is in the group
		group, err := s.groupRepo.FindActiveByUUID(ctx, receiveID)
		if err != nil {
			return nil, errcode.New(errcode.GroupNotFound)
		}
		var members []string
		if err := json.Unmarshal(group.Members, &members); err != nil {
			return nil, err
		}
		inGroup := false
		for _, m := range members {
			if m == sendID {
				inGroup = true
				continue
			}
			recipientIDs = append(recipientIDs, m)
		}
		if !inGroup {
			return nil, errcode.New(errcode.NotInGroup)
		}
	}

	now := time.Now()
	conversationID := buildConversationID(sessionType, sendID, receiveID)

	// Find or create sessions for both parties
	session, err := s.ensureSession(ctx, sendID, receiveID, sessionType, now)
	if err != nil {
		return nil, err
	}

	// Persist the message
	msg := &model.Message{
		Uuid:           xid.New().String(),
		SessionId:      session.Uuid,
		ConversationId: conversationID,
		Type:           msgType,
		Content:        content,
		Url:            url,
		SendId:         sendID,
		ReceiveId:      receiveID,
		Status:         1,
		SendAt:         &now,
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Update session summaries
	digest := content
	if msgType != 0 {
		digest = "[Media Message]"
	}
	_ = s.sessionRepo.UpdateSummaryByUUID(ctx, session.Uuid, digest, &now)

	recipientSessionIDs := make(map[string]string, len(recipientIDs))
	for _, recipientID := range recipientIDs {
		targetSess, err := s.ensureSession(ctx, recipientID, receiveID, sessionType, now)
		if sessionType == 0 {
			targetSess, err = s.ensureSession(ctx, recipientID, sendID, 0, now)
		}
		if err != nil {
			continue
		}
		recipientSessionIDs[recipientID] = targetSess.Uuid
		_ = s.sessionRepo.UpdateSummaryByUUID(ctx, targetSess.Uuid, digest, &now)
		_ = s.sessionRepo.IncrementUnreadCountByUUID(ctx, targetSess.Uuid)
	}

	return &SendMessageResult{
		Message: &MessageInfo{
			Uuid:      msg.Uuid,
			SessionId: session.Uuid,
			Type:      msg.Type,
			Content:   msg.Content,
			Url:       msg.Url,
			SendId:    msg.SendId,
			ReceiveId: msg.ReceiveId,
			Status:    msg.Status,
			SendAt:    now.Format(time.RFC3339),
		},
		RecipientIDs:        recipientIDs,
		RecipientSessionIDs: recipientSessionIDs,
	}, nil
}
