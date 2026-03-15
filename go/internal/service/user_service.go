package service

import (
	"context"
	"errors"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/gorm"
)

const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) EnsureDefaultAdmin(ctx context.Context) error {
	count, err := s.users.Count(ctx)
	if err != nil {
		return apperr.Internal("user_count_failed", "failed to inspect users", err)
	}

	if count > 0 {
		return nil
	}

	passwordHash, err := HashPassword(DefaultAdminPassword)
	if err != nil {
		return apperr.Internal("user_password_hash_failed", "failed to create default admin", err)
	}

	user := model.User{
		Username:     DefaultAdminUsername,
		PasswordHash: passwordHash,
		Role:         model.UserRoleAdmin,
	}

	if err := s.users.Create(ctx, &user); err != nil {
		return apperr.Internal("default_admin_create_failed", "failed to create default admin", err)
	}

	return nil
}

func (s *UserService) List(ctx context.Context) ([]UserRecord, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, apperr.Internal("user_list_failed", "failed to list users", err)
	}

	items := make([]UserRecord, 0, len(users))
	for _, user := range users {
		items = append(items, toUserRecord(user))
	}

	return items, nil
}

func (s *UserService) Create(ctx context.Context, input UserInput) (UserRecord, error) {
	if err := validateUserInput(input.Username, input.Password, input.Role, true); err != nil {
		return UserRecord{}, err
	}

	if _, err := s.findByUsername(ctx, input.Username); err == nil {
		return UserRecord{}, apperr.Conflict("username_exists", "username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserRecord{}, apperr.Internal("user_lookup_failed", "failed to inspect username", err)
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return UserRecord{}, apperr.Internal("user_password_hash_failed", "failed to create user", err)
	}

	user := model.User{
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: passwordHash,
		Role:         input.Role,
	}

	if err := s.users.Create(ctx, &user); err != nil {
		return UserRecord{}, apperr.Internal("user_create_failed", "failed to create user", err)
	}

	return toUserRecord(user), nil
}

func (s *UserService) Update(ctx context.Context, id uint64, input UserInput) (UserRecord, error) {
	if err := validateUserInput(input.Username, input.Password, input.Role, false); err != nil {
		return UserRecord{}, err
	}

	user, err := s.findUser(ctx, id)
	if err != nil {
		return UserRecord{}, err
	}

	existing, err := s.findByUsername(ctx, input.Username)
	if err == nil && existing.ID != user.ID {
		return UserRecord{}, apperr.Conflict("username_exists", "username already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserRecord{}, apperr.Internal("user_lookup_failed", "failed to inspect username", err)
	}

	if user.Role == model.UserRoleAdmin && input.Role != model.UserRoleAdmin {
		adminCount, err := s.users.CountByRole(ctx, model.UserRoleAdmin)
		if err != nil {
			return UserRecord{}, apperr.Internal("user_admin_count_failed", "failed to inspect admin users", err)
		}

		if adminCount <= 1 {
			return UserRecord{}, apperr.Conflict("last_admin_required", "at least one admin user must remain")
		}
	}

	user.Username = strings.TrimSpace(input.Username)
	user.Role = input.Role

	if strings.TrimSpace(input.Password) != "" {
		passwordHash, err := HashPassword(input.Password)
		if err != nil {
			return UserRecord{}, apperr.Internal("user_password_hash_failed", "failed to update user password", err)
		}
		user.PasswordHash = passwordHash
	}

	if err := s.users.Save(ctx, &user); err != nil {
		return UserRecord{}, apperr.Internal("user_update_failed", "failed to update user", err)
	}

	return toUserRecord(user), nil
}

func (s *UserService) Delete(ctx context.Context, id uint64, currentUserID uint64) error {
	if id == currentUserID {
		return apperr.Conflict("cannot_delete_self", "cannot delete the current login user")
	}

	user, err := s.findUser(ctx, id)
	if err != nil {
		return err
	}

	if user.Role == model.UserRoleAdmin {
		adminCount, err := s.users.CountByRole(ctx, model.UserRoleAdmin)
		if err != nil {
			return apperr.Internal("user_admin_count_failed", "failed to inspect admin users", err)
		}

		if adminCount <= 1 {
			return apperr.Conflict("last_admin_required", "at least one admin user must remain")
		}
	}

	if err := s.users.Delete(ctx, id); err != nil {
		return apperr.Internal("user_delete_failed", "failed to delete user", err)
	}

	return nil
}

func (s *UserService) FindRecordByID(ctx context.Context, id uint64) (UserRecord, error) {
	user, err := s.findUser(ctx, id)
	if err != nil {
		return UserRecord{}, err
	}

	return toUserRecord(user), nil
}

func (s *UserService) findUser(ctx context.Context, id uint64) (model.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, apperr.NotFound("user_not_found", "user not found")
		}

		return model.User{}, apperr.Internal("user_lookup_failed", "failed to load user", err)
	}

	return user, nil
}

func (s *UserService) findByUsername(ctx context.Context, username string) (model.User, error) {
	return s.users.FindByUsername(ctx, strings.TrimSpace(username))
}

func toUserRecord(user model.User) UserRecord {
	return UserRecord{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func validateUserInput(username, password string, role model.UserRole, passwordRequired bool) error {
	if strings.TrimSpace(username) == "" {
		return apperr.BadRequest("username_required", "username is required")
	}

	if passwordRequired && strings.TrimSpace(password) == "" {
		return apperr.BadRequest("password_required", "password is required")
	}

	switch role {
	case model.UserRoleAdmin, model.UserRoleMember:
	default:
		return apperr.BadRequest("invalid_user_role", "role must be admin or member")
	}

	return nil
}
