package repository

import (
    "context"
    "errors"

    "wchat/internal/model"

    "gorm.io/gorm"
)

type SessionRepo struct {
    db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
    return &SessionRepo{db: db}
}

func (r *SessionRepo) FindSessionByTarget(
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

func (r *SessionRepo) Create(ctx context.Context, session *model.Session) error {
    return gorm.G[model.Session](r.db).Create(ctx, session)
}

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

func (r *SessionRepo) FindByUUID(ctx context.Context, uuid string) (*model.Session, error) {
    session, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).First(ctx)
    if err != nil {
        return nil, err
    }
    return &session, nil
}

func (r *SessionRepo) DeleteByUUID(ctx context.Context, uuid string) error {
    _, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).Delete(ctx)
    return err
}

func (r *SessionRepo) UpdateField(ctx context.Context, uuid string, field string, value any) error {
    _, err := gorm.G[model.Session](r.db).Where("uuid = ?", uuid).Update(ctx, field, value)
    return err
}
