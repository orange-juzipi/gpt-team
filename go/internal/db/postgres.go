package db

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/config"
	"gpt-team-api/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var databaseNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func Open(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	adminDSN, err := replaceDatabaseName(cfg.PostgresDSN, "postgres")
	if err != nil {
		return nil, err
	}

	appDSN, err := replaceDatabaseName(cfg.PostgresDSN, cfg.DatabaseName)
	if err != nil {
		return nil, err
	}

	if err := ensureDatabase(ctx, adminDSN, cfg.DatabaseName); err != nil {
		return nil, err
	}

	database, err := gorm.Open(postgres.Open(appDSN), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, apperr.Internal("postgres_connect_failed", "failed to connect to postgres", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, apperr.Internal("postgres_db_failed", "failed to create sql db handle", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := database.WithContext(ctx).AutoMigrate(
		&model.Card{},
		&model.CardEvent{},
		&model.Account{},
		&model.MailboxProvider{},
		&model.User{},
	); err != nil {
		return nil, apperr.Internal("postgres_migrate_failed", "failed to auto migrate database", err)
	}

	return database, nil
}

func ensureDatabase(ctx context.Context, adminDSN, name string) error {
	if !databaseNamePattern.MatchString(name) {
		return apperr.BadRequest("invalid_database_name", "database name must contain only letters, numbers, and underscores")
	}

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return apperr.Internal("postgres_admin_connect_failed", "failed to connect to postgres admin database", err)
	}

	var exists int
	if err := adminDB.WithContext(ctx).
		Raw("SELECT 1 FROM pg_database WHERE datname = ?", name).
		Scan(&exists).Error; err != nil {
		return apperr.Internal("postgres_database_check_failed", "failed to check database existence", err)
	}

	if exists == 1 {
		return nil
	}

	statement := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	if err := adminDB.WithContext(ctx).Exec(statement).Error; err != nil {
		return apperr.Internal("postgres_database_create_failed", "failed to create database", err)
	}

	return nil
}

func replaceDatabaseName(dsn, database string) (string, error) {
	if strings.Contains(dsn, "://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return "", apperr.BadRequest("invalid_postgres_dsn", "POSTGRES_DSN must be a valid postgres connection string")
		}

		parsed.Path = "/" + database
		return parsed.String(), nil
	}

	parts := strings.Fields(dsn)
	found := false
	for index, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			parts[index] = "dbname=" + database
			found = true
		}
	}

	if !found {
		parts = append(parts, "dbname="+database)
	}

	if len(parts) == 0 {
		return "", apperr.BadRequest("invalid_postgres_dsn", "POSTGRES_DSN must not be empty")
	}

	return strings.Join(parts, " "), nil
}
