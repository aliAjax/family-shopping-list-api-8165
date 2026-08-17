package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"time"

	"family-shopping-list-api/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func Open(cfg config.Config) (*sql.DB, error) {
	address := net.JoinHostPort(cfg.DBHost, cfg.DBPort)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=UTC&multiStatements=true",
		cfg.DBUser, cfg.DBPassword, address, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func WaitForDB(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("等待 MySQL 就绪超时")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func Migrate(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) NOT NULL PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		version := filepath.Base(file)
		var applied int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("执行迁移 %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("migration applied: %s", version)
	}
	return nil
}

func SeedDemoData(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	demoPasswordHash, err := bcryptHash("demo12345")
	if err != nil {
		return err
	}
	alicePasswordHash, err := bcryptHash("alice12345")
	if err != nil {
		return err
	}
	bobPasswordHash, err := bcryptHash("bob12345")
	if err != nil {
		return err
	}

	users := []struct {
		username string
		nickname string
		hash     string
	}{
		{"demo", "演示用户", demoPasswordHash},
		{"alice", "爱丽丝", alicePasswordHash},
		{"bob", "鲍勃", bobPasswordHash},
	}
	for _, user := range users {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE username = ?`, user.username).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)`,
				user.username, user.hash, user.nickname); err != nil {
				return err
			}
		}
	}

	var demoID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'demo'`).Scan(&demoID); err != nil {
		return err
	}
	var listID uint64
	err = tx.QueryRowContext(ctx, `SELECT id FROM shopping_lists WHERE owner_id = ? AND name = ? LIMIT 1`, demoID, "家庭日常采购").Scan(&listID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO shopping_lists (name, description, owner_id) VALUES (?, ?, ?)`,
			"家庭日常采购", "演示清单：用于验证共享、邀请码和商品编辑", demoID)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		listID = uint64(id)
	}

	ensureMember := func(username string, role string) error {
		var userID uint64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, username).Scan(&userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT IGNORE INTO members (list_id, user_id, role, joined_at) VALUES (?, ?, ?, UTC_TIMESTAMP())`,
			listID, userID, role)
		return err
	}
	if err := ensureMember("demo", "owner"); err != nil {
		return err
	}
	if err := ensureMember("alice", "member"); err != nil {
		return err
	}
	if err := ensureMember("bob", "member"); err != nil {
		return err
	}

	var itemCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM items WHERE list_id = ?`, listID).Scan(&itemCount); err != nil {
		return err
	}
	if itemCount == 0 {
		items := []struct {
			name     string
			quantity int
		}{
			{"牛奶", 2},
			{"鸡蛋", 1},
		}
		for _, item := range items {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO items (list_id, name, quantity, purchased, created_by, updated_by)
				 VALUES (?, ?, ?, 0, ?, ?)`,
				listID, item.name, item.quantity, demoID, demoID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
