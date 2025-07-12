package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type RedisTokenStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewTokenStore(addr string) *RedisTokenStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisTokenStore{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisTokenStore) Set(userID string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	ttl := time.Until(token.Expiry)
	if ttl <= 0 {
		ttl = 30 * time.Minute // fallback TTL
	}

	return r.client.Set(r.ctx, fmt.Sprintf("token:%s", userID), data, ttl).Err()
}

func (r *RedisTokenStore) Get(userID string) (*oauth2.Token, error) {
	val, err := r.client.Get(r.ctx, fmt.Sprintf("token:%s", userID)).Result()
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(val), &token); err != nil {
		return nil, err
	}

	return &token, nil
}
