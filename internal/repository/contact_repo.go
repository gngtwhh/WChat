package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "wchat/internal/model"
)

type ContactRepo struct {
    db *gorm.DB
}

func NewContactRepo(db *gorm.DB) *ContactRepo {
    return &ContactRepo{
        db: db,
    }
}

type ContactDetail struct {
    ContactId string
    Nickname  string
    Avatar    string
    Signature string
    Status    int8
}

func (r *ContactRepo) FindUserContactsWithDetails(
    ctx context.Context, userID string, contactType int8,
) ([]ContactDetail, error) {
    var results []ContactDetail

    err := r.db.WithContext(ctx).Table("contacts c").
        Select("c.contact_id, u.nickname, u.avatar, u.signature, c.status").
        Joins("LEFT JOIN users u ON c.contact_id = u.uuid").
        Where("c.user_id = ? AND c.contact_type = ?", userID, contactType).
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    return results, nil
}

func (r *ContactRepo) CheckContactExists(ctx context.Context, userID, targetID string) (bool, error) {
    _, err := gorm.G[model.Contact](r.db).
        Select("id").
        Where("user_id = ? AND contact_id = ?", userID, targetID).
        First(ctx)

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func (r *ContactRepo) UpdateStatusTx(
    ctx context.Context, userID, targetID string, activeStatus, passiveStatus int8,
) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            err := tx.Model(&model.Contact{}).
                Where("user_id = ? AND contact_id = ?", userID, targetID).
                Update("status", activeStatus).Error
            if err != nil {
                return err
            }

            err = tx.Model(&model.Contact{}).
                Where("user_id = ? AND contact_id = ?", targetID, userID).
                Update("status", passiveStatus).Error
            if err != nil {
                return err
            }

            return nil
        },
    )
}
