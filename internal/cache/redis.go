package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(host string, port int, password string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	RDB = redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
			PoolSize: 100,
		},
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("connect to redis failed: %w", err)
	}
	return nil
}

// SetJSON serialize object to json encoding and store
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, data, expiration).Err()
}

// GetJSON unserialize json encoding to object and return
func GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	data, err := RDB.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(data, dest)
	return true, err
}
