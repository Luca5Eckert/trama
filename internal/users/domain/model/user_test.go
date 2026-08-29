package model_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

func TestNewUserRejectsBlankName(t *testing.T) {
	_, err := model.NewUser("id-1", "   ", time.Now())
	if !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
}

func TestNewUserNormalizesNameAndTime(t *testing.T) {
	location := time.FixedZone("test", -3*60*60)
	now := time.Date(2026, 8, 29, 19, 0, 0, 0, location)

	user, err := model.NewUser("id-1", "  Ada Lovelace  ", now)
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	if user.Name != "Ada Lovelace" {
		t.Fatalf("got name %q", user.Name)
	}
	if user.CreatedAt.Location() != time.UTC || !user.CreatedAt.Equal(now) {
		t.Fatalf("got time %v, want UTC representation of %v", user.CreatedAt, now)
	}
}
