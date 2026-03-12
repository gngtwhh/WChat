package repository

import (
    "context"

    "wchat/internal/model"

    "gorm.io/gorm"
)

type MessageRepo struct {
    db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
    return &MessageRepo{db: db}
}

func (r *MessageRepo) FindMessageList(ctx context.Context, sessionID string, offset, limit int) (
    int64, []model.Message, error,
) {
    total, err := gorm.G[model.Message](r.db).
        Where("session_id = ?", sessionID).
        Count(ctx, "*")
    if err != nil {
        return 0, nil, err
    }

    messages, err := gorm.G[model.Message](r.db).
        Where("session_id = ?", sessionID).
        Order("send_at DESC").
        Offset(offset).Limit(limit).
        Find(ctx)

    return total, messages, err
}

func (r *MessageRepo) FindByUUID(ctx context.Context, uuid string) (*model.Message, error) {
    msg, err := gorm.G[model.Message](r.db).Where("uuid = ?", uuid).First(ctx)
    if err != nil {
        return nil, err
    }
    return &msg, nil
}

func (r *MessageRepo) UpdateStatus(ctx context.Context, uuid string, status int8) error {
    _, err := gorm.G[model.Message](r.db).
        Where("uuid = ?", uuid).
        Update(ctx, "status", status)
    return err
}
