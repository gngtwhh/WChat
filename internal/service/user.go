package service

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "wchat/internal/model"
    "wchat/internal/repository"
    "wchat/pkg/errcode"
)

type UserService struct {
    userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
    user, err := s.userRepo.FindByUUID(ctx, uuid)
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
    return s.userRepo.UpdatesByUUID(ctx, uuid, data)
}

func (s *UserService) GetUserList(ctx context.Context, page, size int, keyword string) (int64, []*model.User, error) {
    offset := (page - 1) * size
    return s.userRepo.FindList(ctx, offset, size, keyword)
}

// TODO: checkAdmin should replace by the auth middleware(JWT)
func (s *UserService) checkAdmin(ctx context.Context, operatorID string) error {
    operator, err := s.userRepo.FindByUUID(ctx, operatorID)
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
    return s.userRepo.UpdateFieldByUUID(ctx, targetUUID, "status", status)
}

// SetUserRole set user role
func (s *UserService) SetUserRole(ctx context.Context, operatorID, targetUUID string, isAdmin int8) error {
    if err := s.checkAdmin(ctx, operatorID); err != nil {
        return err
    }
    return s.userRepo.UpdateFieldByUUID(ctx, targetUUID, "is_admin", isAdmin)
}

func (s *UserService) DeleteUser(ctx context.Context, operatorID, targetUUID string) error {
    if err := s.checkAdmin(ctx, operatorID); err != nil {
        return err
    }

    if operatorID == targetUUID {
        return errcode.NewWithMsg(errcode.ParamError, "不能删除自身账号")
    }
    return s.userRepo.DeleteByUUID(ctx, targetUUID)
}
