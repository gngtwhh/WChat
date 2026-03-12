package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "wchat/internal/model"
)

type ContactApplyRepo struct {
    db *gorm.DB
}

func NewContactApplyRepo(db *gorm.DB) *ContactApplyRepo {
    return &ContactApplyRepo{db: db}
}

// FindReceivedApplies 获取该用户收到的所有申请
// 包含了别人加自己的(ContactType=0)，以及别人加自己群的(ContactType=1)
func (r *ContactApplyRepo) FindReceivedApplies(ctx context.Context, userID string) ([]model.ContactApply, error) {
    var applies []model.ContactApply

    err := r.db.WithContext(ctx).
        Where(
            "(contact_type = 0 AND contact_id = ?) OR (contact_type = 1 AND contact_id IN (SELECT uuid FROM groups WHERE owner_id = ?))",
            userID, userID,
        ).
        Order("last_apply_at DESC").
        Find(&applies).Error

    return applies, err
}

// FindApplyByUUID 查询单条申请
func (r *ContactApplyRepo) FindApplyByUUID(ctx context.Context, uuid string) (*model.ContactApply, error) {
    apply, err := gorm.G[model.ContactApply](r.db).Where("uuid = ?", uuid).First(ctx)
    if err != nil {
        return nil, err
    }
    return &apply, nil
}

// SaveApply 保存或更新申请记录 (如果之前有拒绝的记录，重新发起时要更新时间和状态)
func (r *ContactApplyRepo) SaveApply(ctx context.Context, apply *model.ContactApply) error {
    return r.db.WithContext(ctx).Save(apply).Error
}

// GetApplyByUserAndTarget 查找双方是否已经有过申请记录
func (r *ContactApplyRepo) GetApplyByUserAndTarget(
    ctx context.Context, userID, contactID string, contactType int8,
) (*model.ContactApply, error) {
    apply, err := gorm.G[model.ContactApply](r.db).
        Where("user_id = ? AND contact_id = ? AND contact_type = ?", userID, contactID, contactType).
        First(ctx)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &apply, nil
}

// HandleUserApplyTx 处理好友申请 (更新申请状态 + 建立双向好友关系)
func (r *ContactApplyRepo) HandleUserApplyTx(
    ctx context.Context, applyUUID string, applicantID, targetID string, status int8,
) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            // 更新申请状态
            _, err := gorm.G[model.ContactApply](tx).
                Where("uuid = ?", applyUUID).
                Update(ctx, "status", status)
            if err != nil {
                return err
            }

            // 如果是同意 (status=1)，则建立双向的 Contact 记录
            if status == 1 {
                contacts := []model.Contact{
                    {UserId: applicantID, ContactId: targetID, ContactType: 0, Status: 0}, // A 加 B
                    {UserId: targetID, ContactId: applicantID, ContactType: 0, Status: 0}, // B 加 A
                }
                if err := tx.Create(&contacts).Error; err != nil {
                    return err
                }
            }

            return nil
        },
    )
}
