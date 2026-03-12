package cache

import (
    "context"
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
