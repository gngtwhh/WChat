package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"

	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"
)

type ApplicationInfo struct {
	Uuid        string
	UserId      string
	Nickname    string
	Avatar      string
	ContactId   string
	ContactType int8
	Status      int8
	Message     string
	LastApplyAt string
}

type ApplicationService struct {
	applyRepo   *repository.ContactApplyRepo
	userRepo    *repository.UserRepo
	contactRepo *repository.ContactRepo
	groupRepo   *repository.GroupRepo
}

func NewApplicationService(
	applyRepo *repository.ContactApplyRepo,
	userRepo *repository.UserRepo,
	contactRepo *repository.ContactRepo,
	groupRepo *repository.GroupRepo,
) *ApplicationService {
	return &ApplicationService{
		applyRepo:   applyRepo,
		userRepo:    userRepo,
		contactRepo: contactRepo,
		groupRepo:   groupRepo,
	}
}

func isUserInGroup(group *model.Group, userID string) (bool, error) {
	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		return false, err
	}
	for _, memberUUID := range members {
		if memberUUID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *ApplicationService) GetApplicationList(ctx context.Context, userID string) ([]ApplicationInfo, error) {
	applies, err := s.applyRepo.FindReceivedByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(applies) == 0 {
		return []ApplicationInfo{}, nil
	}

	var applicantIDs []string
	for _, app := range applies {
		applicantIDs = append(applicantIDs, app.UserId)
	}

	users, err := s.userRepo.FindByUUIDs(ctx, applicantIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.Uuid] = u
	}

	infos := make([]ApplicationInfo, 0, len(applies))
	for _, app := range applies {
		applicant := userMap[app.UserId]
		infos = append(
			infos, ApplicationInfo{
				Uuid:        app.Uuid,
				UserId:      app.UserId,
				Nickname:    applicant.Nickname,
				Avatar:      applicant.Avatar,
				ContactId:   app.ContactId,
				ContactType: app.ContactType,
				Status:      app.Status,
				Message:     app.Message,
				LastApplyAt: app.LastApplyAt.Format(time.RFC3339),
			},
		)
	}

	return infos, nil
}

func (s *ApplicationService) SubmitApplication(
	ctx context.Context, applicantID, contactID string, contactType int8, message string,
) error {
	if applicantID == contactID {
		return errcode.New(errcode.CannotAddSelf)
	}

	if contactType == 0 {
		_, err := s.userRepo.FindActiveByUUID(ctx, contactID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.New(errcode.UserNotFound)
			}
			return err
		}

		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, applicantID, contactID)
		if err != nil {
			return err
		}
		if isFriend {
			return errcode.New(errcode.AlreadyFriends)
		}

		// TODO: check whether the target user has blocked the applicant.
	} else {
		group, err := s.groupRepo.FindActiveByUUID(ctx, contactID)
		if err != nil {
			return errcode.New(errcode.GroupNotFound)
		}
		if group.Status == 2 {
			return errcode.New(errcode.GroupDismissed)
		}

		inGroup, err := isUserInGroup(group, applicantID)
		if err != nil {
			return err
		}
		if inGroup {
			return errcode.New(errcode.AlreadyInGroup)
		}

		if group.AddMode == 0 {
			return s.groupRepo.AddMembers(ctx, contactID, []string{applicantID})
		}
	}

	apply, err := s.applyRepo.FindByUserAndContact(ctx, applicantID, contactID, contactType)
	if err != nil {
		return err
	}

	if apply != nil {
		if apply.Status == 0 {
			return errcode.NewWithMsg(errcode.ParamError, "您的申请正在处理中，请耐心等待")
		}
		apply.Status = 0
		apply.Message = message
		apply.LastApplyAt = time.Now()
		return s.applyRepo.Save(ctx, apply)
	}

	newApply := &model.ContactApply{
		Uuid:        xid.New().String(),
		UserId:      applicantID,
		ContactId:   contactID,
		ContactType: contactType,
		Status:      0,
		Message:     message,
		LastApplyAt: time.Now(),
	}
	return s.applyRepo.Save(ctx, newApply)
}

func (s *ApplicationService) HandleApplication(ctx context.Context, operatorID, applyUUID string, status int8) error {
	apply, err := s.applyRepo.FindByUUID(ctx, applyUUID)
	if err != nil {
		return errcode.New(errcode.ApplyNotFound)
	}

	if apply.ContactType == 0 && apply.ContactId != operatorID {
		return errcode.New(errcode.Unauthorized)
	}
	if apply.ContactType == 1 {
		group, err := s.groupRepo.FindActiveByUUID(ctx, apply.ContactId)
		if err != nil {
			return errcode.New(errcode.GroupNotFound)
		}
		if group.OwnerId != operatorID {
			return errcode.New(errcode.Unauthorized)
		}
	}

	if apply.Status != 0 {
		return errcode.New(errcode.ApplyAlreadyHandled)
	}

	if apply.ContactType == 0 {
		isFriend, err := s.contactRepo.ExistsActiveFriendship(ctx, apply.UserId, apply.ContactId)
		if err != nil {
			return err
		}
		if isFriend && status == 1 {
			apply.Status = 1
			if err := s.applyRepo.Save(ctx, apply); err != nil {
				return err
			}
			return errcode.New(errcode.AlreadyFriends)
		}

		return s.applyRepo.HandleUserApply(ctx, applyUUID, apply.UserId, apply.ContactId, status)
	}

	group, err := s.groupRepo.FindActiveByUUID(ctx, apply.ContactId)
	if err != nil {
		return errcode.New(errcode.GroupNotFound)
	}
	if group.Status == 2 {
		return errcode.New(errcode.GroupDismissed)
	}

	inGroup, err := isUserInGroup(group, apply.UserId)
	if err != nil {
		return err
	}
	if status == 1 && inGroup {
		apply.Status = 1
		if err := s.applyRepo.Save(ctx, apply); err != nil {
			return err
		}
		return errcode.New(errcode.AlreadyInGroup)
	}

	if status == 1 {
		if err := s.groupRepo.AddMembers(ctx, apply.ContactId, []string{apply.UserId}); err != nil {
			return err
		}
	}

	apply.Status = status
	return s.applyRepo.Save(ctx, apply)
}
