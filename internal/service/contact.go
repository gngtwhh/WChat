package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wchat/internal/repository"
	"wchat/pkg/errcode"
)

type ContactInfo struct {
	ContactId string
	Nickname  string
	Avatar    string
	Signature string
	Status    int8
}

type ContactService struct {
	contactRepo *repository.ContactRepo
}

func NewContactService(contactRepo *repository.ContactRepo) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
	}
}

func (s *ContactService) GetUserContactList(ctx context.Context, userID string) ([]ContactInfo, error) {
	repoDetails, err := s.contactRepo.FindUserContactsWithDetails(ctx, userID, 0)
	if err != nil {
		return nil, err
	}

	infos := make([]ContactInfo, 0, len(repoDetails))
	for _, d := range repoDetails {
		infos = append(
			infos, ContactInfo{
				ContactId: d.ContactId,
				Nickname:  d.Nickname,
				Avatar:    d.Avatar,
				Signature: d.Signature,
				Status:    d.Status,
			},
		)
	}

	return infos, nil
}

func (s *ContactService) DeleteContact(ctx context.Context, userID, targetID string) error {
	return mapContactActionErr(
		s.contactRepo.DeletePair(ctx, userID, targetID),
		"contact relation changed, delete is not allowed",
	)
}

func (s *ContactService) BlockContact(ctx context.Context, userID, targetID string) error {
	return mapContactActionErr(
		s.contactRepo.BlockPair(ctx, userID, targetID),
		"contact relation changed, block is not allowed",
	)
}

func (s *ContactService) UnblockContact(ctx context.Context, userID, targetID string) error {
	return mapContactActionErr(
		s.contactRepo.UnblockPair(ctx, userID, targetID),
		"contact is not blocked by you, unblock is not allowed",
	)
}

func mapContactActionErr(err error, conflictMsg string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.ContactNotFound)
	}
	if errors.Is(err, repository.ErrContactStateConflict) {
		return errcode.NewWithMsg(errcode.ParamError, conflictMsg)
	}

	return err
}
