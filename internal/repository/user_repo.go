package repository

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"wchat/internal/cache"
	"wchat/internal/model"
	"wchat/pkg/zlog"
)

const (
	userCachePrefix = "user:profile:"
	userCacheTTL    = 24 * time.Hour
)

func userCacheKey(uuid string) string {
	return userCachePrefix + uuid
}

type UserRepo struct {
	db *gorm.DB
}

var ErrUserAlreadyExists = errors.New("user already exists")

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	err := gorm.G[model.User](r.db).Create(ctx, user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindActiveByUUID(ctx context.Context, uuid string) (*model.User, error) {
	key := userCacheKey(uuid)

	var user model.User
	found, err := cache.GetJSON(ctx, key, &user)
	if err != nil {
		zlog.Warn("redis get user cache failed", zap.String("uuid", uuid), zap.Error(err))
	}
	if found {
		return &user, nil
	}

	dbUser, err := gorm.G[model.User](r.db).
		Where("uuid = ?", uuid).
		First(ctx)
	if err != nil {
		return nil, err
	}

	if err := cache.SetJSON(ctx, key, &dbUser, userCacheTTL); err != nil {
		zlog.Warn("redis set user cache failed", zap.String("uuid", uuid), zap.Error(err))
	}

	return &dbUser, nil
}

func (r *UserRepo) FindActiveByTelephone(ctx context.Context, telephone string) (*model.User, error) {
	user, err := gorm.G[model.User](r.db).
		Where("telephone = ?", telephone).
		First(ctx)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) CheckActiveTelephoneExists(ctx context.Context, telephone string) (bool, error) {
	_, err := gorm.G[model.User](r.db).
		Select("id").
		Where("telephone = ?", telephone).
		First(ctx)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *UserRepo) UpdateLastOnlineAt(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Update("last_online_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepo) UpdateLastOfflineAt(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Update("last_offline_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepo) UpdateProfileByUUID(ctx context.Context, uuid string, data map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) UpdateStatusByUUID(ctx context.Context, uuid string, status int8) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) UpdateIsAdminByUUID(ctx context.Context, uuid string, isAdmin int8) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Update("is_admin", isAdmin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) UpdateTelephoneByUUID(ctx context.Context, uuid, newTelephone string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Update("telephone", newTelephone)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) MarkDeletionRequested(ctx context.Context, uuid string, requestedAt, deleteAfter time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Updates(
			map[string]any{
				"delete_requested_at": requestedAt,
				"delete_after":        deleteAfter,
			},
		)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) ClearDeletionRequested(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", uuid).
		Updates(
			map[string]any{
				"delete_requested_at": nil,
				"delete_after":        nil,
			},
		)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	cache.RDB.Del(ctx, userCacheKey(uuid))
	return nil
}

func (r *UserRepo) FindList(ctx context.Context, offset, limit int, keyword string) (int64, []*model.User, error) {
	if keyword == "" {
		return 0, make([]*model.User, 0), nil
	}

	chain := gorm.G[model.User](r.db).
		Where("nickname LIKE ? OR telephone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	total, err := chain.Count(ctx, "*")
	if err != nil {
		return 0, nil, err
	}

	users, err := chain.Offset(offset).Limit(limit).Find(ctx)
	if err != nil {
		return 0, nil, err
	}

	result := make([]*model.User, len(users))
	for idx := range users {
		result[idx] = &users[idx]
	}

	return total, result, nil
}

func (r *UserRepo) FindByUUIDs(ctx context.Context, uuids []string) ([]model.User, error) {
	if len(uuids) == 0 {
		return make([]model.User, 0), nil
	}

	result := make([]model.User, 0, len(uuids))
	var missUUIDs []string

	for _, uuid := range uuids {
		var user model.User
		found, err := cache.GetJSON(ctx, userCacheKey(uuid), &user)
		if err != nil {
			zlog.Warn("redis get user cache failed", zap.String("uuid", uuid), zap.Error(err))
		}
		if found {
			result = append(result, user)
		} else {
			missUUIDs = append(missUUIDs, uuid)
		}
	}

	if len(missUUIDs) == 0 {
		return result, nil
	}

	dbUsers, err := gorm.G[model.User](r.db).
		Where("uuid IN ?", missUUIDs).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	for i := range dbUsers {
		if err := cache.SetJSON(ctx, userCacheKey(dbUsers[i].Uuid), &dbUsers[i], userCacheTTL); err != nil {
			zlog.Warn("redis set user cache failed", zap.String("uuid", dbUsers[i].Uuid), zap.Error(err))
		}
	}

	result = append(result, dbUsers...)
	return result, nil
}
