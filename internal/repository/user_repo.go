package repository

import (
    "context"
    "errors"
    "time"

    "gorm.io/gorm"

    "wchat/internal/model"
)

type UserRepo struct {
    db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{
        db: db,
    }
}

func (r *UserRepo) CheckByTelephone(ctx context.Context, telephone string) (bool, error) {
    _, err := gorm.G[model.User](r.db).
        Select("id").
        Where("telephone = ?", telephone).
        First(ctx)

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, nil
        }
        return false, err
    }

    return true, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
    return gorm.G[model.User](r.db).Create(ctx, user)
}

func (r *UserRepo) FindByTelephone(ctx context.Context, telephone string) (*model.User, error) {
    user, err := gorm.G[model.User](r.db).
        Where("telephone = ?", telephone).
        First(ctx)

    if err != nil {
        return nil, err
    }

    return &user, nil
}

func (r *UserRepo) UpdateLastOnlineAt(ctx context.Context, uuid string) error {
    _, err := gorm.G[model.User](r.db).
        Where("uuid = ?", uuid).
        Update(ctx, "last_online_at", time.Now())
    return err
}

func (r *UserRepo) UpdateLastOfflineAt(ctx context.Context, uuid string) error {
    _, err := gorm.G[model.User](r.db).
        Where("uuid = ?", uuid).
        Update(ctx, "last_offline_at", time.Now())
    return err
}

func (r *UserRepo) FindByUUID(ctx context.Context, uuid string) (*model.User, error) {
    user, err := gorm.G[model.User](r.db).
        Where("uuid = ?", uuid).
        First(ctx)

    if err != nil {
        return nil, err
    }

    return &user, nil
}

func (r *UserRepo) UpdatesByUUID(ctx context.Context, uuid string, data map[string]any) error {
    // _, err := gorm.G[model.User](r.db).
    //     Where("uuid = ?", uuid).
    //     Updates(ctx, data)
    // return err
    err := r.db.WithContext(ctx).
        Model(&model.User{}).
        Where("uuid = ?", uuid).
        Updates(data).Error

    return err
}

func (r *UserRepo) UpdateFieldByUUID(ctx context.Context, uuid string, field string, value any) error {
    _, err := gorm.G[model.User](r.db).
        Where("uuid = ?", uuid).
        Update(ctx, field, value)
    return err
}

func (r *UserRepo) FindList(ctx context.Context, offset, limit int, keyword string) (int64, []*model.User, error) {
    q := r.db.WithContext(ctx).Model(new(model.User))

    if keyword != "" {
        q = q.Where("nickname LIKE ? OR telephone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
    }

    var total int64
    if err := q.Count(&total).Error; err != nil {
        return 0, nil, err
    }

    var users []*model.User
    if err := q.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return 0, nil, err
    }

    return total, users, nil
}

func (r *UserRepo) DeleteByUUID(ctx context.Context, uuid string) error {
    _, err := gorm.G[model.User](r.db).
        Where("uuid = ?", uuid).
        Delete(ctx)
    return err
}

func (r *UserRepo) FindByUUIDs(ctx context.Context, uuids []string) ([]model.User, error) {
    if len(uuids) == 0 {
        return make([]model.User, 0), nil
    }

    return gorm.G[model.User](r.db).
        Where("uuid IN ?", uuids).
        Find(ctx)
}
