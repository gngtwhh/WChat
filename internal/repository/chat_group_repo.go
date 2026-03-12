package repository

import (
    "context"
    "encoding/json"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    "wchat/internal/model"
)

type GroupRepo struct {
    db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
    return &GroupRepo{db: db}
}

func (r *GroupRepo) CreateGroupTx(ctx context.Context, group *model.Group, userIDs []string) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            // 创建群组记录
            if err := tx.Create(group).Error; err != nil {
                return err
            }

            // 批量创建群成员的 Contact 记录 (ContactType = 1 代表群组)
            contacts := make([]model.Contact, 0, len(userIDs))
            for _, uid := range userIDs {
                contacts = append(
                    contacts, model.Contact{
                        UserId:      uid,
                        ContactId:   group.Uuid,
                        ContactType: 1,
                        Status:      0, // 正常
                    },
                )
            }
            return tx.Create(&contacts).Error
        },
    )
}

func (r *GroupRepo) FindJoinedGroups(ctx context.Context, userID string) ([]model.Group, error) {
    var groups []model.Group
    err := r.db.WithContext(ctx).Table("groups").
        Joins("JOIN contacts c ON c.contact_id = groups.uuid").
        // c.contact_type = 1 (群聊), c.status = 0 (在群内), groups.status = 0 (群组未解散)
        Where("c.user_id = ? AND c.contact_type = 1 AND c.status = 0 AND groups.status = 0", userID).
        Find(&groups).Error

    return groups, err
}

func (r *GroupRepo) FindByUUID(ctx context.Context, uuid string) (*model.Group, error) {
    group, err := gorm.G[model.Group](r.db).
        Where("uuid = ?", uuid).
        First(ctx)
    if err != nil {
        return nil, err
    }
    return &group, nil
}

func (r *GroupRepo) FindByUUIDs(ctx context.Context, uuids []string) ([]model.Group, error) {
    if len(uuids) == 0 {
        return make([]model.Group, 0), nil
    }

    return gorm.G[model.Group](r.db).
        Where("uuid IN ?", uuids).
        Find(ctx)
}

func (r *GroupRepo) UpdateGroup(ctx context.Context, groupUUID string, data map[string]any) error {
    return r.db.WithContext(ctx).Model(&model.Group{}).
        Where("uuid = ?", groupUUID).Updates(data).Error
}

func (r *GroupRepo) DismissGroupTx(ctx context.Context, groupUUID string) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            // 修改群状态为 2 (解散)
            if err := tx.Model(&model.Group{}).Where("uuid = ?", groupUUID).Update("status", 2).Error; err != nil {
                return err
            }
            // 将所有群成员的联系记录状态改为 6 (退出群聊/解散)
            return tx.Model(&model.Contact{}).
                Where("contact_id = ? AND contact_type = 1", groupUUID).
                Update("status", 6).Error
        },
    )
}

func (r *GroupRepo) AddMembersTx(ctx context.Context, groupUUID string, newMemberIDs []string) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            var group model.Group
            if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where(
                "uuid = ?", groupUUID,
            ).First(&group).Error; err != nil {
                return err
            }

            var currentMembers []string
            _ = json.Unmarshal(group.Members, &currentMembers)

            memberMap := make(map[string]bool)
            for _, m := range currentMembers {
                memberMap[m] = true
            }

            var actuallyAdded []string
            for _, m := range newMemberIDs {
                if !memberMap[m] {
                    currentMembers = append(currentMembers, m)
                    actuallyAdded = append(actuallyAdded, m)
                    memberMap[m] = true
                }
            }

            if len(actuallyAdded) == 0 {
                return nil // 没人需要被添加 (都在群里了)
            }

            newMembersJSON, _ := json.Marshal(currentMembers)
            if err := tx.Model(&group).Updates(
                map[string]any{
                    "members":    newMembersJSON,
                    "member_cnt": len(currentMembers),
                },
            ).Error; err != nil {
                return err
            }

            var contacts []model.Contact
            for _, uid := range actuallyAdded {
                contacts = append(
                    contacts, model.Contact{
                        UserId:      uid,
                        ContactId:   groupUUID,
                        ContactType: 1,
                        Status:      0,
                    },
                )
            }
            return tx.Create(&contacts).Error
        },
    )
}

func (r *GroupRepo) RemoveMemberTx(ctx context.Context, groupUUID, memberUUID string, contactStatus int8) error {
    return r.db.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            var group model.Group
            if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where(
                "uuid = ?", groupUUID,
            ).First(&group).Error; err != nil {
                return err
            }

            var currentMembers []string
            _ = json.Unmarshal(group.Members, &currentMembers)

            // 过滤掉被移除的人
            var newMembers []string
            for _, m := range currentMembers {
                if m != memberUUID {
                    newMembers = append(newMembers, m)
                }
            }

            if len(newMembers) == len(currentMembers) {
                return nil // 没找到这个人，无需操作
            }

            // 更新 Group
            newMembersJSON, _ := json.Marshal(newMembers)
            if err := tx.Model(&group).Updates(
                map[string]any{
                    "members":    newMembersJSON,
                    "member_cnt": len(newMembers),
                },
            ).Error; err != nil {
                return err
            }

            // 2. 更新该成员的 Contact 状态 (例如: 6=主动退出，7=被踢出)
            return tx.Model(&model.Contact{}).
                Where("user_id = ? AND contact_id = ? AND contact_type = 1", memberUUID, groupUUID).
                Update("status", contactStatus).Error
        },
    )
}
