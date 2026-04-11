package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

var ErrContactStateConflict = errors.New("contact state conflict")

func (r *ContactRepo) FindUserContactsWithDetails(
	ctx context.Context, userID string, contactType int8,
) ([]ContactDetail, error) {
	var results []ContactDetail

	err := r.db.WithContext(ctx).Table("contacts c").
		Select("c.contact_id, u.nickname, u.avatar, u.signature, c.status").
		Joins("LEFT JOIN users u ON c.contact_id = u.uuid AND u.deleted_at IS NULL").
		Where("c.user_id = ? AND c.contact_type = ? AND c.deleted_at IS NULL", userID, contactType).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ContactRepo) ExistsByUserAndContact(ctx context.Context, userID, targetID string) (bool, error) {
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

func (r *ContactRepo) ExistsActiveFriendship(ctx context.Context, userID, targetID string) (bool, error) {
	_, err := gorm.G[model.Contact](r.db).
		Select("id").
		Where("user_id = ? AND contact_id = ? AND contact_type = 0 AND status = 0", userID, targetID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ContactRepo) DeleteByUserUUID(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? OR contact_id = ?", uuid, uuid).
		Delete(&model.Contact{}).Error
}

func (r *ContactRepo) DeletePair(ctx context.Context, userID, targetID string) error {
	return r.transitionPair(
		ctx, userID, targetID,
		model.ContactStatusNormal, model.ContactStatusNormal,
		model.ContactStatusDeleted, model.ContactStatusDeletedByPeer,
	)
}

func (r *ContactRepo) BlockPair(ctx context.Context, userID, targetID string) error {
	return r.transitionPair(
		ctx, userID, targetID,
		model.ContactStatusNormal, model.ContactStatusNormal,
		model.ContactStatusBlocked, model.ContactStatusBlockedByPeer,
	)
}

func (r *ContactRepo) UnblockPair(ctx context.Context, userID, targetID string) error {
	return r.transitionPair(
		ctx, userID, targetID,
		model.ContactStatusBlocked, model.ContactStatusBlockedByPeer,
		model.ContactStatusNormal, model.ContactStatusNormal,
	)
}

func (r *ContactRepo) transitionPair(
	ctx context.Context, userID, targetID string, currentActiveStatus, currentPassiveStatus, activeStatus, passiveStatus int8,
) error {
	return r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var contacts []model.Contact
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("user_id", "contact_id", "status").
				Where(
					"contact_type = ? AND ((user_id = ? AND contact_id = ?) OR (user_id = ? AND contact_id = ?))",
					model.ContactTypeUser, userID, targetID, targetID, userID,
				).
				Find(&contacts).Error
			if err != nil {
				return err
			}
			if len(contacts) != 2 {
				return gorm.ErrRecordNotFound
			}

			var (
				activeCurrent  int8
				passiveCurrent int8
				activeFound    bool
				passiveFound   bool
			)
			for i := range contacts {
				contact := contacts[i]
				switch {
				case contact.UserId == userID && contact.ContactId == targetID:
					activeCurrent = contact.Status
					activeFound = true
				case contact.UserId == targetID && contact.ContactId == userID:
					passiveCurrent = contact.Status
					passiveFound = true
				}
			}
			if !activeFound || !passiveFound {
				return gorm.ErrRecordNotFound
			}
			if activeCurrent == activeStatus && passiveCurrent == passiveStatus {
				return nil
			}
			if activeCurrent != currentActiveStatus || passiveCurrent != currentPassiveStatus {
				return ErrContactStateConflict
			}

			result := tx.Model(&model.Contact{}).
				Where(
					"user_id = ? AND contact_id = ? AND contact_type = ? AND status = ?",
					userID, targetID, model.ContactTypeUser, currentActiveStatus,
				).
				Update("status", activeStatus)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ErrContactStateConflict
			}

			result = tx.Model(&model.Contact{}).
				Where(
					"user_id = ? AND contact_id = ? AND contact_type = ? AND status = ?",
					targetID, userID, model.ContactTypeUser, currentPassiveStatus,
				).
				Update("status", passiveStatus)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ErrContactStateConflict
			}

			return nil
		},
	)
}
