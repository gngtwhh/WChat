package service

import (
	"context"
	"errors"
	"wchat/internal/cache"
	"wchat/internal/model"
	"wchat/internal/repository"
	"wchat/pkg/errcode"
	"wchat/pkg/utils"
	"wchat/pkg/zlog"

	"github.com/rs/xid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepo
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, errcode.New(errcode.TokenMissing)
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil, errcode.New(errcode.TokenExpired)
	}

	blacklisted, err := cache.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return nil, errcode.New(errcode.ServerError)
	}
	if blacklisted {
		return nil, errcode.New(errcode.TokenInvalid)
	}

	user, err := s.userRepo.FindActiveByUUID(ctx, claims.Uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.TokenInvalid)
		}
		return nil, errcode.New(errcode.ServerError)
	}

	if user.Status == 1 {
		return nil, errcode.New(errcode.AccountDisabled)
	}
	if user.DeleteAfter != nil {
		return nil, errcode.New(errcode.AccountPendingDeletion)
	}

	return user, nil
}

func (s *AuthService) Register(ctx context.Context, telephone, password, nickname string) error {
	exists, err := s.userRepo.CheckActiveTelephoneExists(ctx, telephone)
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
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return errcode.New(errcode.UserExists)
		}
		return err
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, telephone, password string) (*model.User, string, error) {
	user, err := s.userRepo.FindActiveByTelephone(ctx, telephone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Warn("login failed: user not found", zap.String("telephone", telephone))
			return nil, "", errcode.New(errcode.AuthFailed)
		}
		zlog.Error("login failed: db error", zap.String("telephone", telephone), zap.Error(err))
		return nil, "", errcode.New(errcode.ServerError)
	}

	if !utils.CheckPassword(user.Password, password) {
		zlog.Warn(
			"login failed: password mismatch",
			zap.String("telephone", telephone),
			zap.String("stored_hash", user.Password),
			zap.Int("hash_len", len(user.Password)),
		)
		return nil, "", errcode.New(errcode.AuthFailed)
	}

	// TODO: redis检查用户是否被注销?---多端登录
	// 注销账号无法直接影响到jwt，需要考虑使用缓存记录uuid？

	if user.Status == 1 {
		return nil, "", errcode.New(errcode.AccountDisabled)
	}
	if user.DeleteAfter != nil {
		return nil, "", errcode.New(errcode.AccountPendingDeletion)
	}

	token, err := utils.GenToken(user.Uuid, user.IsAdmin)
	if err != nil {
		return nil, "", err
	}

	// 异步更新最后登录时间 (可选，不阻塞登录流程)
	go s.userRepo.UpdateLastOnlineAt(context.Background(), user.Uuid)

	return user, token, nil
}

func (s *AuthService) Logout(ctx context.Context, uuid string, token string) error {
	if err := s.userRepo.UpdateLastOfflineAt(ctx, uuid); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil
	}
	if err := cache.AddTokenToBlacklist(ctx, token, claims.ExpiresAt.Time); err != nil {
		zlog.Error("failed to add token to blacklist", zap.Error(err))
		return err
	}
	return nil
}
