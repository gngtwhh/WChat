package cache

import (
	"context"
	"fmt"
	"time"
)

const jwtBlacklistPrefix = "jwt:blacklist:"

func AddTokenToBlacklist(ctx context.Context, token string, expireAt time.Time) error {
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return nil // token 已过期，无需加入黑名单
	}
	key := jwtBlacklistPrefix + token
	return RDB.Set(ctx, key, 1, ttl).Err()
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := jwtBlacklistPrefix + token
	n, err := RDB.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check token blacklist failed: %w", err)
	}
	return n > 0, nil
}
