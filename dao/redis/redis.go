/**
 *@Author: IronHuang
 *@Date: 2020/8/18 9:52 下午
**/

package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var rdb *redis.Client

func Init() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
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
