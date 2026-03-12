package service

import (
    "context"

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

    // Maps repo-DTO to service-DTO to enforce strict layer decoupling and prevent import cycles;
    // to eliminate this boilerplate,
    // the struct could alternatively be sunk into the bottom-most 'model' package.
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

func (s *ContactService) UpdateBiDirectionalStatus(
    ctx context.Context, userID, targetID string, activeStatus, passiveStatus int8,
) error {
    exists, err := s.contactRepo.CheckContactExists(ctx, userID, targetID)
    if err != nil {
        return err
    }
    if !exists {
        return errcode.New(errcode.ContactNotFound)
    }

    err = s.contactRepo.UpdateStatusTx(ctx, userID, targetID, activeStatus, passiveStatus)
    if err != nil {
        return err
    }

    return nil
}
