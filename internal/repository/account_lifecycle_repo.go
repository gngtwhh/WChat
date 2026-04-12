package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wchat/internal/cache"
	"wchat/internal/model"
)

func (r *UserRepo) FindPurgeCandidates(ctx context.Context, now time.Time, limit int) ([]model.User, error) {
	return gorm.G[model.User](r.db).
		Where("delete_after IS NOT NULL AND delete_after <= ?", now).
		Limit(limit).
		Find(ctx)
}

func (r *UserRepo) FinalizeDeleteByUUID(ctx context.Context, uuid string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Updates(
			map[string]any{
				"telephone":           fmt.Sprintf("deleted:%s", uuid),
				"nickname":            "已注销用户",
				"email":               "",
				"avatar":              "",
				"signature":           "",
				"birthday":            "",
				"password":            "",
				"is_admin":            0,
				"delete_requested_at": nil,
				"delete_after":        nil,
				"deleted_at":          now,
			},
		)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if cache.RDB != nil {
		cache.RDB.Del(ctx, userCacheKey(uuid))
	}
	return nil
}
