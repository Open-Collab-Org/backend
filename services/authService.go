package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/open-collaboration/server/models"
	"gorm.io/gorm"
	"time"
)

var ErrInvalidSessionKey = errors.New("invalid session key")

type AuthService struct {
	Db           *gorm.DB
	Redis        *redis.Client
	UsersService *UsersService
}

// Check if a session exists and, if it does, return the session's user.
// Returns ErrInvalidSessionKey if the session does not exist.
func (s *AuthService) AuthenticateSession(ctx context.Context, sessionKey string) (*models.User, error) {
	logger := log.FromContext(ctx)

	logger.Debug("Checking for session in redis")

	redisKey := sessionRedisKey(sessionKey)
	userId, err := s.Redis.Get(ctx, redisKey).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Debug("Session does not exist")

			return nil, ErrInvalidSessionKey
		} else {
			logger.WithError(err).Error("Failed to check for session in redis")

			return nil, err
		}
	}

	logger.Debug("Session found, getting user from db")

	user, err := s.UsersService.GetUser(ctx, uint(userId))
	if err != nil {
		logger.
			WithError(err).
			WithFields(log.Fields{
				"sessionKey": sessionKey,
				"userId":     userId,
			}).
			Error("Failed to get the session's user, perhaps the user was deleted from the database?")

		return nil, err
	}

	return user, nil
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
	var sessionsSet []string

	redisKey := sessionInvertedIndexRedisKey(userId)
	err := s.Redis.GetSet(ctx, redisKey, &sessionsSet).Err()
	if err != nil {
		return err
	}

	for i, key := range sessionsSet {
		sessionsSet[i] = sessionRedisKey(key)
	}

	err = s.Redis.Del(ctx, sessionsSet...).Err()
	if err != nil {
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
