package service

import (
	"context"
	"errors"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	users    *repository.UserRepository
	sessions *SessionManager
}

func NewAuthService(users *repository.UserRepository, sessions *SessionManager) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
	}
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthLoginResult, error) {
	if strings.TrimSpace(input.Username) == "" || input.Password == "" {
		return AuthLoginResult{}, apperr.BadRequest("invalid_login_payload", "username and password are required")
	}

	user, err := s.users.FindByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthLoginResult{}, apperr.Unauthorized("invalid_credentials", "username or password is incorrect")
		}

		return AuthLoginResult{}, apperr.Internal("user_lookup_failed", "failed to load user", err)
	}

	if err := CheckPassword(user.PasswordHash, input.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return AuthLoginResult{}, apperr.Unauthorized("invalid_credentials", "username or password is incorrect")
		}

		return AuthLoginResult{}, apperr.Internal("password_check_failed", "failed to verify password", err)
	}

	token, err := s.sessions.Issue(user.ID)
	if err != nil {
		return AuthLoginResult{}, err
	}

	return AuthLoginResult{
		User:  toUserRecord(user),
		Token: token,
	}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, token string) (UserRecord, error) {
	userID, err := s.sessions.Verify(token)
	if err != nil {
		return UserRecord{}, err
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserRecord{}, apperr.Unauthorized("invalid_session", "login required")
		}

		return UserRecord{}, apperr.Internal("user_lookup_failed", "failed to load user", err)
	}

	return toUserRecord(user), nil
}
