package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
	userpostgres "github.com/Luca5Eckert/trama/internal/users/infrastructure/adapter/postgres"
)

func TestUserRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	id := fmt.Sprintf("repository-test-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	}()

	repository := userpostgres.NewUserRepository(pool)
	user := model.User{ID: id, Name: "Ada Lovelace", CreatedAt: time.Now().UTC().Truncate(time.Microsecond)}

	if err := repository.Create(ctx, user); err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err := repository.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if found != user {
		t.Fatalf("got %#v, want %#v", found, user)
	}

	users, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	seen := false
	for _, candidate := range users {
		if candidate.ID == id {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("created user %q was not returned by List", id)
	}

	_, err = repository.GetByID(ctx, id+"-missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
