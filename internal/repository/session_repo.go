package repository

import (
	"context"
	"errors"
	"time"

	"wchat/internal/model"

	"gorm.io/gorm"
)

// SessionRepo handles session persistence operations.
type SessionRepo struct {
	db *gorm.DB
}

// NewSessionRepo creates a new SessionRepo with the given database handle.
func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// findUnscoped finds a session by user ID, target ID and session type, including soft deleted records.
func (r *SessionRepo) findUnscoped(
	ctx context.Context, userID, targetID string, sessionType int8,
) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND target_id = ? AND session_type = ?", userID, targetID, sessionType).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// restoreIfDeleted restores a soft-deleted session and resets its summary fields.
func (r *SessionRepo) restoreIfDeleted(
	ctx context.Context, session *model.Session, lastMessageAt *time.Time,
) (*model.Session, error) {
	if session.DeletedAt.Time.IsZero() {
		return session, nil
	}

	if err := r.db.WithContext(ctx).
		Unscoped().
		Model(&model.Session{}).
		Where("id = ?", session.ID).
		Updates(
			map[string]any{
				"deleted_at":      nil,
				"unread_count":    0,
				"last_message":    "",
				"last_message_at": lastMessageAt,
				"is_top":          0,
			},
		).Error; err != nil {
		return nil, err
	}

	session.DeletedAt = gorm.DeletedAt{}
	session.UnreadCount = 0
	session.LastMessage = ""
	session.LastMessageAt = lastMessageAt
	session.IsTop = 0
	return session, nil
}

// FindOrCreate returns an existing session identified by target or creates a new one if none exists.
func (r *SessionRepo) FindOrCreate(
	ctx context.Context, userID, targetID string, sessionType int8, sessionUUID string, lastMessageAt *time.Time,
) (*model.Session, error) {
	session, err := r.findUnscoped(ctx, userID, targetID, sessionType)
	if err == nil {
		return r.restoreIfDeleted(ctx, session, lastMessageAt)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	session = &model.Session{
		Uuid:          sessionUUID,
		UserId:        userID,
		TargetId:      targetID,
		SessionType:   sessionType,
		LastMessageAt: lastMessageAt,
	}
	if err := gorm.G[model.Session](r.db).Create(ctx, session); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			session, readErr := r.findUnscoped(ctx, userID, targetID, sessionType)
			if readErr != nil {
				return nil, readErr
			}
			return r.restoreIfDeleted(ctx, session, lastMessageAt)
		}
		return nil, err
	}
	return session, nil
}

// DeleteByUserUUID deletes all sessions belonging to the specified user.
func (r *SessionRepo) DeleteByUserUUID(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", uuid).
		Delete(&model.Session{}).Error
}

// FindSessionList returns a paginated list of sessions for the user.
func (r *SessionRepo) FindSessionList(ctx context.Context, userID string, offset, limit int) (
	int64, []model.Session, error,
) {
	total, err := gorm.G[model.Session](r.db).
		Where("user_id = ?", userID).
		Count(ctx, "*")
	if err != nil {
		return 0, nil, err
	}

	sessions, err := gorm.G[model.Session](r.db).
		Where("user_id = ?", userID).
		Order("is_top DESC, last_message_at DESC").
		Offset(offset).Limit(limit).
		Find(ctx)

	return total, sessions, err
}

// FindActiveByUUID returns an active session by its UUID.
func (r *SessionRepo) FindActiveByUUID(ctx context.Context, uuid string) (*model.Session, error) {
	session, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).First(ctx)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveByUserTarget returns an active session matching the user, target, and type.
func (r *SessionRepo) FindActiveByUserTarget(
	ctx context.Context, userID, targetID string, sessionType int8,
) (*model.Session, error) {
	session, err := gorm.G[model.Session](r.db).
		Where("user_id = ? AND target_id = ? AND session_type = ?", userID, targetID, sessionType).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// SoftDeleteByUUID soft deletes the session with the specified UUID.
func (r *SessionRepo) SoftDeleteByUUID(ctx context.Context, uuid string) error {
	_, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).Delete(ctx)
	return err
}

// UpdateTopStatusByUUID updates the pinned status of a session.
func (r *SessionRepo) UpdateTopStatusByUUID(ctx context.Context, uuid string, isTop int8) error {
	_, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).Update(ctx, "is_top", isTop)
	return err
}

// ClearUnreadCountByUUID resets the unread count for the session.
func (r *SessionRepo) ClearUnreadCountByUUID(ctx context.Context, uuid string) error {
	_, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).Update(ctx, "unread_count", 0)
	return err
}

// IncrementUnreadCountByUUID increments the unread count for the session.
func (r *SessionRepo) IncrementUnreadCountByUUID(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("uuid = ?", uuid).
		UpdateColumn("unread_count", gorm.Expr("unread_count + ?", 1)).Error
}

// DecrementUnreadCountByUUID decrements the unread count but does not go below zero.
func (r *SessionRepo) DecrementUnreadCountByUUID(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("uuid = ? AND unread_count > 0", uuid).
		UpdateColumn("unread_count", gorm.Expr("unread_count - ?", 1)).Error
}

// UpdateSummaryByUUID updates the last message summary fields for the session.
func (r *SessionRepo) UpdateSummaryByUUID(
	ctx context.Context, uuid string, lastMessage string, lastMessageAt *time.Time,
) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("uuid = ?", uuid).
		Updates(
			map[string]any{
				"last_message":    lastMessage,
				"last_message_at": lastMessageAt,
			},
		).Error
}
