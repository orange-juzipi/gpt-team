package service

import (
	"context"
	"fmt"
	"testing"

	"gpt-team-api/internal/model"
	"gpt-team-api/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnsureDefaultAdminCreatesAdminOnEmptyDatabase(t *testing.T) {
	t.Parallel()

	database, err := openTestUserDatabase(t)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := database.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	repo := repository.NewUserRepository(database)
	svc := NewUserService(repo)

	if err := svc.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure default admin: %v", err)
	}

	user, err := repo.FindByUsername(context.Background(), DefaultAdminUsername)
	if err != nil {
		t.Fatalf("find default admin: %v", err)
	}

	if user.Role != model.UserRoleAdmin {
		t.Fatalf("expected admin role, got %s", user.Role)
	}

	if err := CheckPassword(user.PasswordHash, DefaultAdminPassword); err != nil {
		t.Fatalf("expected default admin password to match: %v", err)
	}
}

func TestEnsureDefaultAdminDoesNotCreateDuplicateAdmin(t *testing.T) {
	t.Parallel()

	database, err := openTestUserDatabase(t)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := database.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	repo := repository.NewUserRepository(database)
	svc := NewUserService(repo)

	existingHash, err := HashPassword("existing-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := repo.Create(context.Background(), &model.User{
		Username:     "existing-admin",
		PasswordHash: existingHash,
		Role:         model.UserRoleAdmin,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := svc.EnsureDefaultAdmin(context.Background()); err != nil {
		t.Fatalf("ensure default admin: %v", err)
	}

	count, err := repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count users: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected existing users to remain unchanged, got %d users", count)
	}
}

func openTestUserDatabase(t *testing.T) (*gorm.DB, error) {
	t.Helper()

	return gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{},
	)
}
