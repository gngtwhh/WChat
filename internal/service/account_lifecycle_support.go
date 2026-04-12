package service

import (
	"context"
	"time"

	"wchat/internal/repository"
)

func forceDeleteAccount(ctx context.Context, txManager *repository.TxManager, uuid string) error {
	return txManager.InTx(
		ctx,
		func(repos *repository.TxRepos) error {
			if err := repos.Group.DismissOwnedGroups(ctx, uuid); err != nil {
				return err
			}
			if err := repos.Group.RemoveUserFromJoinedGroups(ctx, uuid); err != nil {
				return err
			}
			if err := repos.Contact.DeleteByUserUUID(ctx, uuid); err != nil {
				return err
			}
			if err := repos.Session.DeleteByUserUUID(ctx, uuid); err != nil {
				return err
			}
			if err := repos.ContactApply.DeleteByUserUUID(ctx, uuid); err != nil {
				return err
			}
			return repos.User.FinalizeDeleteByUUID(ctx, uuid)
		},
	)
}

func purgeExpiredAccounts(
	ctx context.Context, userRepo *repository.UserRepo, txManager *repository.TxManager, now time.Time, limit int,
) (int, error) {
	if limit <= 0 {
		limit = 50
	}

	users, err := userRepo.FindPurgeCandidates(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	if len(users) == 0 {
		return 0, nil
	}

	purged := 0
	for i := range users {
		if err := forceDeleteAccount(ctx, txManager, users[i].Uuid); err != nil {
			return purged, err
		}
		purged++
	}

	return purged, nil
}
