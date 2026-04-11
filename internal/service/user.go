package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wchat/internal/cache"
	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"
	"wchat/pkg/utils"
)

const accountDeletionCooldown = 7 * 24 * time.Hour

type UserService struct {
	userRepo  *repository.UserRepo
	txManager *repository.TxManager
}

func NewUserService(
	userRepo *repository.UserRepo,
	txManager *repository.TxManager,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	user, err := s.userRepo.FindActiveByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.UserNotFound)
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, uuid string, data map[string]any) error {
	if len(data) == 0 {
		return nil
	}
	err := s.userRepo.UpdateProfileByUUID(ctx, uuid, data)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.UserNotFound)
	}
	return err
}

func (s *UserService) RequestAccountDeletion(ctx context.Context, uuid, password, token string) error {
	user, err := s.userRepo.FindActiveByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.UserNotFound)
		}
		return err
	}
	if !utils.CheckPassword(user.Password, password) {
		return errcode.New(errcode.InvalidPassword)
	}
	if user.DeleteAfter != nil {
		return nil
	}

	now := time.Now()
	if err := s.userRepo.MarkDeletionRequested(ctx, uuid, now, now.Add(accountDeletionCooldown)); err != nil {
		return err
	}
	if err := s.userRepo.UpdateLastOfflineAt(ctx, uuid); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if token == "" {
		return nil
	}
	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil
	}
	return cache.AddTokenToBlacklist(ctx, token, claims.ExpiresAt.Time)
}

func (s *UserService) CancelAccountDeletion(ctx context.Context, uuid, password string) error {
	user, err := s.userRepo.FindActiveByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.UserNotFound)
		}
		return err
	}
	if !utils.CheckPassword(user.Password, password) {
		return errcode.New(errcode.InvalidPassword)
	}
	if user.DeleteAfter == nil {
		return nil
	}

	return s.userRepo.ClearDeletionRequested(ctx, uuid)
}

func (s *UserService) ChangeTelephone(ctx context.Context, uuid, password, newTelephone, verifyCode string) error {
	user, err := s.userRepo.FindActiveByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.UserNotFound)
		}
		return err
	}
	if !utils.CheckPassword(user.Password, password) {
		return errcode.New(errcode.InvalidPassword)
	}
	if user.DeleteAfter != nil {
		return errcode.New(errcode.AccountPendingDeletion)
	}
	if verifyCode == "" {
		return errcode.NewWithMsg(errcode.ParamError, "验证码不能为空")
	}
	if newTelephone == user.Telephone {
		return nil
	}
	// exists, err := s.userRepo.CheckActiveTelephoneExists(ctx, newTelephone)
	// if err != nil {
	// 	return err
	// }
	// if exists {
	// 	return errcode.New(errcode.UserExists)
	// }
	err = s.userRepo.UpdateTelephoneByUUID(ctx, uuid, newTelephone)
	if errors.Is(err, repository.ErrUserAlreadyExists) || errors.Is(err, gorm.ErrDuplicatedKey) {
		return errcode.New(errcode.UserExists)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.UserNotFound)
	}
	return err
}

func (s *UserService) GetUserList(ctx context.Context, page, size int, keyword string) (int64, []*model.User, error) {
	offset := (page - 1) * size
	return s.userRepo.FindList(ctx, offset, size, keyword)
}

// TODO: checkAdmin should replace by the auth middleware(JWT)
func (s *UserService) checkAdmin(ctx context.Context, operatorID string) error {
	operator, err := s.userRepo.FindActiveByUUID(ctx, operatorID)
	if err != nil {
		return errcode.New(errcode.UserNotFound)
	}
	if operator.IsAdmin != 1 {
		return errcode.New(errcode.Unauthorized)
	}
	return nil
}

// SetUserStatus enable or disable user
func (s *UserService) SetUserStatus(ctx context.Context, operatorID, targetUUID string, status int8) error {
	if err := s.checkAdmin(ctx, operatorID); err != nil {
		return err
	}
	err := s.userRepo.UpdateStatusByUUID(ctx, targetUUID, status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.UserNotFound)
	}
	return err
}

// SetUserRole set user role
func (s *UserService) SetUserRole(ctx context.Context, operatorID, targetUUID string, isAdmin int8) error {
	if err := s.checkAdmin(ctx, operatorID); err != nil {
		return err
	}
	err := s.userRepo.UpdateIsAdminByUUID(ctx, targetUUID, isAdmin)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.UserNotFound)
	}
	return err
}

// DeleteUser directly delete the target user account
// admin role required
func (s *UserService) DeleteUser(ctx context.Context, operatorID, targetUUID string) error {
	if err := s.checkAdmin(ctx, operatorID); err != nil {
		return err
	}

	if operatorID == targetUUID {
		return errcode.NewWithMsg(errcode.ParamError, "不能删除自身账号")
	}
	if err := forceDeleteAccount(ctx, s.txManager, targetUUID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.UserNotFound)
		}
		return err
	}

	return nil
}
