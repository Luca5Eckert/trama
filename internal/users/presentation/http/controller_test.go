package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/users/application/command"
	"github.com/Luca5Eckert/trama/internal/users/application/query"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
	userhttp "github.com/Luca5Eckert/trama/internal/users/presentation/http"
)

type repositoryFake struct{ users map[string]model.User }
func newRepositoryFake() *repositoryFake { return &repositoryFake{users: make(map[string]model.User)} }
func (repo *repositoryFake) Create(_ context.Context, user model.User) error { repo.users[user.ID] = user; return nil }
func (repo *repositoryFake) GetByID(_ context.Context, id string) (model.User, error) { user, ok := repo.users[id]; if !ok { return model.User{}, domain.ErrNotFound }; return user, nil }
func (repo *repositoryFake) List(context.Context) ([]model.User, error) { users := make([]model.User, 0, len(repo.users)); for _, user := range repo.users { users = append(users, user) }; return users, nil }
type idFake struct{ id string }
func (fake idFake) NewID() (string, error) { return fake.id, nil }
type clockFake struct{ now time.Time }
func (fake clockFake) Now() time.Time { return fake.now }

func newHandler(repository *repositoryFake) http.Handler {
	createUser := command.NewCreateUser(repository, idFake{id: "user-1"}, clockFake{now: time.Date(2026, 8, 29, 22, 0, 0, 0, time.UTC)})
	controller := userhttp.NewUserController(createUser, query.NewGetUserByID(repository), query.NewListUsers(repository))
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)
	return mux
}

func TestCreateUser(t *testing.T) {
	handler := newHandler(newRepositoryFake())
	request := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(`{"name":"Ada Lovelace"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated { t.Fatalf("got status %d, body=%s", response.Code, response.Body.String()) }
	if response.Header().Get("Location") != "/v1/users/user-1" { t.Fatalf("got Location %q", response.Header().Get("Location")) }
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil { t.Fatalf("decode response: %v", err) }
	if body["id"] != "user-1" || body["name"] != "Ada Lovelace" { t.Fatalf("unexpected body %#v", body) }
}

func TestCreateUserRejectsInvalidJSON(t *testing.T) {
	handler := newHandler(newRepositoryFake())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(`{"name":`)))
	if response.Code != http.StatusBadRequest { t.Fatalf("got status %d", response.Code) }
}

func TestCreateUserRejectsInvalidName(t *testing.T) {
	handler := newHandler(newRepositoryFake())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(`{"name":"   "}`)))
	if response.Code != http.StatusUnprocessableEntity { t.Fatalf("got status %d, body=%s", response.Code, response.Body.String()) }
	var body struct { Error struct { Code string `json:"code"` } `json:"error"` }
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil { t.Fatalf("decode error: %v", err) }
	if body.Error.Code != "invalid_name" { t.Fatalf("got error code %q", body.Error.Code) }
}

func TestGetUserNotFound(t *testing.T) {
	handler := newHandler(newRepositoryFake())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/users/missing", nil))
	if response.Code != http.StatusNotFound { t.Fatalf("got status %d, body=%s", response.Code, response.Body.String()) }
}

func TestGetUser(t *testing.T) {
	repository := newRepositoryFake()
	repository.users["user-1"] = model.User{ID: "user-1", Name: "Ada", CreatedAt: time.Now().UTC()}
	handler := newHandler(repository)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/users/user-1", nil))
	if response.Code != http.StatusOK { t.Fatalf("got status %d, body=%s", response.Code, response.Body.String()) }
}
