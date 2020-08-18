/**
 *@Author: IronHuang
 *@Date: 2020/8/18 9:52 下午
**/

package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"go_web_cli/settings"
)

var rdb *redis.Client

func Init(cfg *settings.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Db,
	})
	_, err = rdb.Ping().Result()
	if err != nil {
		zap.L().Error("connect RDB failed", zap.Error(err))
		return
	}
	return
}

func Close() {
	_ = rdb.Close()
}
