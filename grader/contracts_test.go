package grader_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	invHandler "family-shopping-list-api/internal/invite/handler"
	invRepo "family-shopping-list-api/internal/invite/repository"
	invService "family-shopping-list-api/internal/invite/service"
	itemHandler "family-shopping-list-api/internal/item/handler"
	itemRepo "family-shopping-list-api/internal/item/repository"
	itemService "family-shopping-list-api/internal/item/service"
	listHandler "family-shopping-list-api/internal/list/handler"
	listRepo "family-shopping-list-api/internal/list/repository"
	listService "family-shopping-list-api/internal/list/service"
	memberHandler "family-shopping-list-api/internal/member/handler"
	memberRepo "family-shopping-list-api/internal/member/repository"
	memberService "family-shopping-list-api/internal/member/service"
	userRepo "family-shopping-list-api/internal/user/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func requireEmptyArray(t *testing.T, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal collection: %v", err)
	}
	if string(encoded) != "[]" {
		t.Fatalf("empty collection encoded as %s, want []", encoded)
	}
}

func TestEmptyCollectionsUseJSONArray(t *testing.T) {
	t.Run("items", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT i.id").WillReturnRows(sqlmock.NewRows([]string{
			"id", "list_id", "name", "quantity", "purchased", "created_by", "updated_by",
			"created_by_name", "updated_by_name", "created_at", "updated_at",
		}))
		items, err := itemRepo.New(db).ListByList(context.Background(), 1)
		if err != nil {
			t.Fatalf("list items: %v", err)
		}
		requireEmptyArray(t, items)
	})

	t.Run("lists", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT l.id").WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "owner_id", "owner_name", "member_count", "created_at", "updated_at",
		}))
		lists, err := listRepo.New(db).ListByUser(context.Background(), 1)
		if err != nil {
			t.Fatalf("list shopping lists: %v", err)
		}
		requireEmptyArray(t, lists)
	})

	t.Run("members", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT m.id").WillReturnRows(sqlmock.NewRows([]string{
			"id", "list_id", "user_id", "role", "username", "nickname", "joined_at", "created_at", "updated_at",
		}))
		members, err := memberRepo.New(db).List(context.Background(), 1)
		if err != nil {
			t.Fatalf("list members: %v", err)
		}
		requireEmptyArray(t, members)
	})

	t.Run("invites", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT i.id").WillReturnRows(sqlmock.NewRows([]string{
			"id", "list_id", "code", "created_by", "creator_name", "max_uses", "used_count",
			"expires_at", "active", "created_at", "updated_at",
		}))
		invites, err := invRepo.New(db).ListByList(context.Background(), 1)
		if err != nil {
			t.Fatalf("list invites: %v", err)
		}
		requireEmptyArray(t, invites)
	})
}

func TestDatabaseErrorsRemainInspectable(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT id").WillReturnError(context.DeadlineExceeded)
		_, err := userRepo.New(db).GetByID(context.Background(), 1)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("database error chain lost: %v", err)
		}
	})
	t.Run("list", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT l.id").WillReturnError(context.DeadlineExceeded)
		_, err := listRepo.New(db).GetByID(context.Background(), 1)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("database error chain lost: %v", err)
		}
	})
	t.Run("item", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT i.id").WillReturnError(context.DeadlineExceeded)
		_, err := itemRepo.New(db).GetByID(context.Background(), 1, 1)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("database error chain lost: %v", err)
		}
	})
	t.Run("invite", func(t *testing.T) {
		db, mock := newMockDB(t)
		mock.ExpectQuery("SELECT i.id").WillReturnError(context.DeadlineExceeded)
		_, err := invRepo.New(db).GetByCode(context.Background(), "CODE")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("database error chain lost: %v", err)
		}
	})

}

func TestCanceledCollectionRequestsStopDatabaseWork(t *testing.T) {
	type handlerCase struct {
		columns []string
		build   func(*sql.DB) http.HandlerFunc
		path    string
	}
	cases := map[string]handlerCase{
		"items": {
			columns: []string{"id", "list_id", "name", "quantity", "purchased", "created_by", "updated_by", "created_by_name", "updated_by_name", "created_at", "updated_at"},
			build: func(db *sql.DB) http.HandlerFunc {
				return itemHandler.New(itemService.New(itemRepo.New(db))).List
			},
			path: "/api/v1/lists/1/items",
		},
		"lists": {
			columns: []string{"id", "name", "description", "owner_id", "owner_name", "member_count", "created_at", "updated_at"},
			build: func(db *sql.DB) http.HandlerFunc {
				return listHandler.New(listService.New(listRepo.New(db), memberRepo.New(db))).List
			},
			path: "/api/v1/lists",
		},
		"members": {
			columns: []string{"id", "list_id", "user_id", "role", "username", "nickname", "joined_at", "created_at", "updated_at"},
			build: func(db *sql.DB) http.HandlerFunc {
				return memberHandler.New(memberService.New(memberRepo.New(db))).List
			},
			path: "/api/v1/lists/1/members",
		},
		"invites": {
			columns: []string{"id", "list_id", "code", "created_by", "creator_name", "max_uses", "used_count", "expires_at", "active", "created_at", "updated_at"},
			build: func(db *sql.DB) http.HandlerFunc {
				return invHandler.New(invService.New(invRepo.New(db), memberRepo.New(db))).List
			},
			path: "/api/v1/lists/1/invites",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db, mock := newMockDB(t)
			mock.ExpectQuery("SELECT").WillDelayFor(250 * time.Millisecond).WillReturnRows(sqlmock.NewRows(tc.columns))
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil).WithContext(ctx)
			if name != "lists" {
				req.SetPathValue("list_id", "1")
			}
			timer := time.AfterFunc(20*time.Millisecond, cancel)
			defer timer.Stop()
			started := time.Now()
			tc.build(db)(httptest.NewRecorder(), req)
			if elapsed := time.Since(started); elapsed > 150*time.Millisecond {
				t.Fatalf("request cancellation took %s to reach database work", elapsed)
			}
		})
	}
}

func TestCollectionQueriesCloseRowsAfterDecodeFailure(t *testing.T) {
	now := time.Now()
	cases := map[string]struct {
		columns []string
		values  []driver.Value
		call    func(*sql.DB) error
	}{
		"items": {
			columns: []string{"id", "list_id", "name", "quantity", "purchased", "created_by", "updated_by", "created_by_name", "updated_by_name", "created_at", "updated_at"},
			values:  []driver.Value{"bad", 1, "milk", 1, false, 1, 1, "A", "A", now, now},
			call:    func(db *sql.DB) error { _, err := itemRepo.New(db).ListByList(context.Background(), 1); return err },
		},
		"lists": {
			columns: []string{"id", "name", "description", "owner_id", "owner_name", "member_count", "created_at", "updated_at"},
			values:  []driver.Value{"bad", "weekly", "", 1, "A", 1, now, now},
			call:    func(db *sql.DB) error { _, err := listRepo.New(db).ListByUser(context.Background(), 1); return err },
		},
		"members": {
			columns: []string{"id", "list_id", "user_id", "role", "username", "nickname", "joined_at", "created_at", "updated_at"},
			values:  []driver.Value{"bad", 1, 1, "member", "a", "A", now, now, now},
			call:    func(db *sql.DB) error { _, err := memberRepo.New(db).List(context.Background(), 1); return err },
		},
		"invites": {
			columns: []string{"id", "list_id", "code", "created_by", "creator_name", "max_uses", "used_count", "expires_at", "active", "created_at", "updated_at"},
			values:  []driver.Value{"bad", 1, "CODE", 1, "A", 5, 0, nil, true, now, now},
			call:    func(db *sql.DB) error { _, err := invRepo.New(db).ListByList(context.Background(), 1); return err },
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db, mock := newMockDB(t)
			rows := sqlmock.NewRows(tc.columns).AddRow(tc.values...).AddRow(tc.values...)
			mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
			if err := tc.call(db); err == nil {
				t.Fatal("expected row decoding to fail")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("query resources were not released: %v", err)
			}
		})
	}
}
