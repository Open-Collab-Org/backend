package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"time"
)

var ErrInvalidSessionToken = errors.New("invalid session key")

type AuthService struct {
	Db           *gorm.DB
	Redis        *redis.Client
	UsersService *UsersService
}

// Check if a session exists and, if it does, return the session's user.
// Returns ErrInvalidSessionToken if the session does not exist.
func (s *AuthService) AuthenticateSession(ctx context.Context, sessionKey string) (uint, error) {
	logger := log.FromContext(ctx)

	logger.Debug("Checking for session in redis")

	redisKey := sessionRedisKey(sessionKey)
	userId, err := s.Redis.Get(ctx, redisKey).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Debug("Session does not exist")

			return 0, ErrInvalidSessionToken
		} else {
			logger.WithError(err).Error("Failed to check for session in redis")

			return 0, err
		}
	}

	logger.Debug("Session is valid")

	return uint(userId), nil
}

// Create a session key for a user. The session key will last 30 days.
func (s *AuthService) CreateSession(ctx context.Context, userId uint) (string, error) {
	logger := log.FromContext(ctx)

	logger.
		WithField("userId", userId).
		Debug("Creating session key")

	// Using uuid is as session key is safe here because it uses the rand
	// package to get random numbers, which in turn uses the rand package, which
	// is cryptographically safe.
	sessionKey, err := uuid.NewV4()
	if err != nil {
		logger.WithError(err).Error("Failed to generate a session key")

		return "", err
	}

	// 1 month
	keyDuration := time.Hour * 24 * 30

	err = s.Redis.Watch(ctx, func(tx *redis.Tx) error {
		err = s.Redis.Set(ctx, sessionRedisKey(sessionKey.String()), userId, keyDuration).Err()
		if err != nil {
			return err
		}

		err = s.Redis.SAdd(ctx, sessionInvertedIndexRedisKey(userId), sessionKey).Err()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.WithError(err).Error("Failed to store session key: redis transaction failed")

		return "", err
	}

	return sessionKey.String(), nil
}

// Invalidate (delete) all sessions of a user.
func (s *AuthService) InvalidateSessions(ctx context.Context, userId uint) error {
	logger := log.FromContext(ctx).WithField("userId", userId)

	logger.Debug("Invalidating all sessions of user")

	var sessionsSet []string

	redisKey := sessionInvertedIndexRedisKey(userId)
	err := s.Redis.GetSet(ctx, redisKey, &sessionsSet).Err()
	if err != nil {
		logger.WithError(err).Error("Failed to get all of a user's session tokens")
		return err
	}

	for i, key := range sessionsSet {
		sessionsSet[i] = sessionRedisKey(key)
	}

	keysToDelete := append([]string{}, sessionsSet...)
	keysToDelete = append(keysToDelete, redisKey)

	err = s.Redis.Del(ctx, keysToDelete...).Err()
	if err != nil {
		logger.WithError(err).Error("Failed to delete session token keys")
		return err
	}

	return nil
}

// Maps a session key to a user id.
func sessionRedisKey(sessionKey string) string {
	return fmt.Sprintf("session:%s:user.id", sessionKey)
}

// Maps a user id to a set of session keys.
//
// It's an inverted index of sessionRedisKey.
func sessionInvertedIndexRedisKey(userId uint) string {
	return fmt.Sprintf("user:%d:session.keys", userId)
}
