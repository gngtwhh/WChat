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

func (r *MessageRepo) FindMessageListByConversation(ctx context.Context, conversationID string, offset, limit int) (
	int64, []model.Message, error,
) {
	total, err := gorm.G[model.Message](r.db).
		Where("conversation_id = ?", conversationID).
		Count(ctx, "*")
	if err != nil {
		return 0, nil, err
	}

	messages, err := gorm.G[model.Message](r.db).
		Where("conversation_id = ?", conversationID).
		Order("send_at DESC").
		Offset(offset).Limit(limit).
		Find(ctx)

	return total, messages, err
}

func (r *MessageRepo) FindActiveByUUID(ctx context.Context, uuid string) (*model.Message, error) {
	msg, err := gorm.G[model.Message](r.db).Where("uuid = ?", uuid).First(ctx)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepo) Create(ctx context.Context, msg *model.Message) error {
	return gorm.G[model.Message](r.db).Create(ctx, msg)
}

func (r *MessageRepo) MarkRecalledByUUID(ctx context.Context, uuid string) error {
	_, err := gorm.G[model.Message](r.db).
		Where("uuid = ?", uuid).
		Update(ctx, "status", 2)
	return err
}
