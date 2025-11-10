package app

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var Rdb *redis.Client

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     Conf.Redis.Host,
		Password: Conf.Redis.Pass,
		DB:       Conf.Redis.DB,
	})
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Redis连接失败", err)
	}
}
