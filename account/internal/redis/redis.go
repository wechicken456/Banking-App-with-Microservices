package redis

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}
type (
	singleClient  struct{ *redis.Client }
	clusterClient struct{ *redis.ClusterClient }
)

func (c *singleClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.Client.Get(ctx, key)
}

func (c *singleClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.Client.Set(ctx, key, value, expiration)
}

func (c *singleClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return c.Client.Del(ctx, keys...)
}

func (c *clusterClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.ClusterClient.Get(ctx, key)
}

func (c *clusterClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.ClusterClient.Set(ctx, key, value, expiration)
}

func (c *clusterClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return c.ClusterClient.Del(ctx, keys...)
}

var Client RedisClient

func Init(ctx context.Context) error {
	mode := os.Getenv("REDIS_MODE")
	if mode == "" {
		mode = "single"
	}

	password := os.Getenv("REDIS_PASSWORD")

	var addrs []string
	if mode == "single" {
		host := os.Getenv("REDIS_SINGLE_ADDR")
		if host == "" {
			host = "redis-single"
		}

		port := os.Getenv("REDIS_SINGLE_PORT")
		if port == "" {
			port = "6379"
		}
		addrs = []string{fmt.Sprintf("%s:%s", host, port)}
	} else {
		clusterAddrs := os.Getenv("REDIS_CLUSTER_ADDRS")
		if clusterAddrs == "" {
			clusterAddrs = "redis-node1:6380,redis-node2:6381,redis-node3:6382"
		}
		addrs = strings.Split(clusterAddrs, ",")
	}

	var err error
	switch mode {
	case "single":
		client := redis.NewClient(&redis.Options{
			Addr:     addrs[0],
			Password: password,
			DB:       0,
		})
		if err = client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis single ping failed: %w", err)
		}
		Client = &singleClient{client}
	case "cluster":
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: password,
		})
		if err = client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis cluster ping failed: %w", err)
		}
		Client = &clusterClient{client}
	default:
		return fmt.Errorf("invalid REDIS_MODE: %s", mode)
	}
	return nil
}
