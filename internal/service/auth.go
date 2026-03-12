package service

import (
    "context"
    "wchat/internal/model"
    "wchat/internal/repository"
    "wchat/pkg/errcode"
    "wchat/pkg/utils"

    "github.com/rs/xid"
)

type AuthService struct {
    userRepo *repository.UserRepo
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
    return &AuthService{
        userRepo: userRepo,
    }
}

func (s *AuthService) Register(ctx context.Context, telephone, password, nickname string) error {
    exists, err := s.userRepo.CheckByTelephone(ctx, telephone)
    if err != nil {
        return err
    }
    if exists {
        return errcode.New(errcode.UserExists)
    }
    hashedPassword, err := utils.HashPassword(password)
    if err != nil {
        return err
    }

    user := &model.User{
        Uuid:      xid.New().String(),
        Telephone: telephone,
        Password:  hashedPassword,
        Nickname:  nickname,
    }
    if err := s.userRepo.Create(ctx, user); err != nil {
        return err
    }
    return nil
}

func (s *AuthService) Login(ctx context.Context, telephone, password string) (*model.User, string, error) {
    user, err := s.userRepo.FindByTelephone(ctx, telephone)
    if err != nil {
        return nil, "", errcode.New(errcode.AuthFailed)
    }

    if !utils.CheckPassword(password, user.Password) {
        return nil, "", errcode.New(errcode.AuthFailed)
    }

    if user.Status == 1 {
        return nil, "", errcode.New(errcode.AccountDisabled)
    }

    token, err := utils.GenToken(user.Uuid, user.IsAdmin)
    if err != nil {
        return nil, "", err
    }

    // 异步更新最后登录时间 (可选，不阻塞登录流程)
    go s.userRepo.UpdateLastOnlineAt(context.Background(), user.Uuid)

    return user, token, nil
}

func (s *AuthService) Logout(ctx context.Context, uuid string) error {
    if err := s.userRepo.UpdateLastOfflineAt(ctx, uuid); err != nil {
        return err
    }
    // TODO: set JWT blacklist in Redis
    return nil
}
