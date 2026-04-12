package service

import (
	"context"
	"time"

	"wchat/internal/repository"
)

type AccountLifecycleService struct {
	userRepo  *repository.UserRepo
	txManager *repository.TxManager
}

func NewAccountLifecycleService(
	userRepo *repository.UserRepo,
	txManager *repository.TxManager,
) *AccountLifecycleService {
	return &AccountLifecycleService{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

func (s *AccountLifecycleService) PurgeExpiredAccounts(ctx context.Context, now time.Time) error {
	const purgeBatchSize = 50
	for {
		purged, err := purgeExpiredAccounts(ctx, s.userRepo, s.txManager, now, purgeBatchSize)
		if err != nil {
			return err
		}
		if purged < purgeBatchSize {
			return nil
		}
	}
}
