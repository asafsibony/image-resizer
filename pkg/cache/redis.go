package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	rdb    *redis.Client
	ctx    context.Context
	logger *logrus.Logger
}

func NewRedisClient(ctx context.Context, logger *logrus.Logger, redisHost string, redisPassword string, redisDB int) *RedisClient {
	return &RedisClient{
		rdb: redis.NewClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPassword,
			DB:       redisDB,
		}),
		ctx:    ctx,
		logger: logger,
	}
}

func (c *RedisClient) Set(key string, value interface{}) (string, error) {
	res, err := c.rdb.Set(c.ctx, key, value, 0).Result()
	if err != nil {
		c.logger.Errorf("%+v\n", err)
		return "", err
	}
	return res, nil
}

func (c *RedisClient) SetNX(key string, value string) (bool, error) {
	res, err := c.rdb.SetNX(c.ctx, key, value, 0).Result()
	if err != nil {
		c.logger.Errorf("%+v\n", err)
		return false, err
	}
	return res, nil
}

func (c *RedisClient) Get(key string) (string, error) {
	res, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		c.logger.Errorf("%+v\n", err)
		return "", err
	}
	return res, nil
}

func (c *RedisClient) GetInt64(key string) (int64, error) {
	res, err := c.rdb.Get(c.ctx, key).Int64()
	if err != nil {
		c.logger.Errorf("%+v\n", err)
		return -1, err
	}
	return res, nil
}

func (c *RedisClient) Delete(key string) (int64, error) {
	res, err := c.rdb.Del(c.ctx, key).Result()
	if err != nil {
		c.logger.Errorf("%+v\n", err)
		return -1, err
	}
	return res, nil
}
