package repository

import (
	"context"

	"gorm.io/gorm"
)

type TxRepos struct {
	User         *UserRepo
	Group        *GroupRepo
	Contact      *ContactRepo
	ContactApply *ContactApplyRepo
	Session      *SessionRepo
	Message      *MessageRepo
}

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) InTx(ctx context.Context, fn func(repos *TxRepos) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(
			&TxRepos{
				User:         NewUserRepo(tx),
				Group:        NewGroupRepo(tx),
				Contact:      NewContactRepo(tx),
				ContactApply: NewContactApplyRepo(tx),
				Session:      NewSessionRepo(tx),
				Message:      NewMessageRepo(tx),
			},
		)
	})
}
