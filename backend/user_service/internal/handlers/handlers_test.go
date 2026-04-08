package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pkg/roles"
	"testing"
	"user_service/internal/handlers"
	"user_service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(method, path string, handlerFn gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Handle(method, path, handlerFn)
	return r
}

func TestCreateUser(t *testing.T) {
	validGUID := uuid.New()
	tests := []struct {
		name       string
		body       any
		srvFn      func(ctx context.Context, user models.User) (uuid.UUID, error)
		wantStatus int
		wantGUID   uuid.UUID
	}{
		{
			name: "успешное создание",
			body: models.User{
				Login:    "login",
				Password: "password",
			},
			srvFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return validGUID, nil
			},
			wantStatus: http.StatusCreated,
			wantGUID:   validGUID,
		},
		{
			name:       "невалидный json",
			body:       "invalid kto? ya invalid? A mozet ti?",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Логин уже занят",
			body: models.User{Login: "login", Password: "password"},
			srvFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return uuid.Nil, models.ErrLoginBusy
			},
			wantStatus: http.StatusConflict,
			wantGUID:   uuid.Nil,
		},
		{
			name: "Ошибка регистрации",
			body: models.User{Login: "login", Password: "password"},
			srvFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return uuid.Nil, models.ErrRegistrationFailed
			},
			wantStatus: http.StatusInternalServerError,
			wantGUID:   uuid.Nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			srv := &mockUserService{
				CreateUserFn: test.srvFn,
			}

			h := handlers.NewHandlers(srv)

			r := setupRouter(http.MethodPost, "/user", h.CreateUser)

			body, _ := json.Marshal(test.body)

			req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)

			if test.wantGUID != uuid.Nil {
				guid := struct {
					GUID uuid.UUID `json:"guid"`
				}{}
				assert.NoError(t, json.NewDecoder(w.Body).Decode(&guid))
				assert.Equal(t, test.wantGUID, guid.GUID)
			}
		})
	}
}

func TestGetUserByGUID(t *testing.T) {
	validGUID := uuid.New()
	validLogin := "valid_login"

	tests := []struct {
		name       string
		guidParam  string
		srvFn      func(ctx context.Context, guid uuid.UUID) (models.User, error)
		wantStatus int
	}{
		{
			name:      "Пользователь найден",
			guidParam: validGUID.String(),
			srvFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{GUID: validGUID, Login: validLogin}, nil
			},
			wantStatus: http.StatusOK,
		},

		{
			name:       "Невалидный GUID",
			guidParam:  "Invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "GUID не найден",
			guidParam: validGUID.String(),
			srvFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{}, models.ErrGUIDNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "Неизвестная ошибка",
			guidParam: validGUID.String(),
			srvFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{}, models.ErrUnknown
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := &mockUserService{
				GetUserByGUIDFn: test.srvFn,
			}

			h := handlers.NewHandlers(srv)

			router := setupRouter(http.MethodGet, "/user/:guid", h.GetUserByGUID)

			req := httptest.NewRequest(http.MethodGet, "/user/"+test.guidParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)
			if w.Code == http.StatusOK {
				user := models.User{}

				assert.NoError(t, json.NewDecoder(w.Body).Decode(&user))
				assert.Equal(t, validLogin, user.Login)
			}
		})
	}
}

func TestGetUserByLogin(t *testing.T) {
	validLogin := "valid_login"
	tests := []struct {
		name       string
		login      string
		srvFn      func(ctx context.Context, login string) (models.User, error)
		wantStatus int
	}{
		{
			name:  "Пользователь найден",
			login: validLogin,
			srvFn: func(ctx context.Context, login string) (models.User, error) {
				return models.User{Login: validLogin}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "Login не найден",
			login: "invalid_login",
			srvFn: func(ctx context.Context, login string) (models.User, error) {
				return models.User{}, models.ErrLoginNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "Неизвестная ошибка",
			login: validLogin,
			srvFn: func(ctx context.Context, login string) (models.User, error) {
				return models.User{}, models.ErrUnknown
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := &mockUserService{
				GetUserByLoginFn: test.srvFn,
			}

			h := handlers.NewHandlers(srv)

			router := setupRouter(http.MethodGet, "/user", h.GetUserByLogin)

			url := "/user?login=" + test.login
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)

			if w.Code == http.StatusOK {
				user := models.User{}

				assert.NoError(t, json.NewDecoder(w.Body).Decode(&user))
				assert.Equal(t, validLogin, user.Login)
			}
		})
	}
}

func TestVerifyUser(t *testing.T) {
	validGUID := uuid.New()
	tests := []struct {
		name       string
		body       any
		srvFn      func(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error)
		wantStatus int
	}{
		{
			name: "Успешная аунтентификация",
			body: models.User{
				Password: "valid_password",
				Login:    "valid_Login",
			},
			srvFn: func(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error) {
				return validGUID, roles.RoleUser, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Невалидный json",
			body:       "Invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Ошибка аунтентификации",
			body: models.User{
				Login:    "valid",
				Password: "valid",
			},
			srvFn: func(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error) {
				return uuid.Nil, roles.RoleUser, models.ErrAuthenticationFailed
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Неизвестная ошибка",
			body: models.User{
				Login:    "valid",
				Password: "valid",
			},
			srvFn: func(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error) {
				return uuid.Nil, roles.RoleUser, models.ErrUnknown
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := &mockUserService{
				VerifyUserFn: test.srvFn,
			}

			h := handlers.NewHandlers(srv)

			body, _ := json.Marshal(test.body)

			r := setupRouter(http.MethodPost, "/user/auth", h.VerifyUser)
			req := httptest.NewRequest(http.MethodPost, "/user/auth", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)

			if w.Code == http.StatusOK {
				user := models.User{}

				assert.NoError(t, json.NewDecoder(w.Body).Decode(&user))
				assert.Equal(t, validGUID, user.GUID)
			}
		})
	}
}
func TestSetRole(t *testing.T) {
	validGUID := uuid.New()
	validRole := struct {
		Role roles.Role `json:"role"`
	}{
		Role: roles.RoleUser,
	}
	tests := []struct {
		name       string
		guid       uuid.UUID
		role       any
		srvFn      func(ctx context.Context, guid uuid.UUID, role roles.Role) error
		wantStatus int
	}{
		{
			name: "успешная смена роли",
			guid: validGUID,
			role: validRole,
			srvFn: func(ctx context.Context, guid uuid.UUID, role roles.Role) error {
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Невалидная роль",
			guid:       validGUID,
			role:       "Invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Неизвестная ошибка сервиса",
			guid: validGUID,
			role: validRole,
			srvFn: func(ctx context.Context, guid uuid.UUID, role roles.Role) error {
				return models.ErrUnknown
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := &mockUserService{
				SetRoleFn: test.srvFn,
			}

			h := handlers.NewHandlers(srv)

			router := setupRouter(http.MethodPatch, "/user/:guid/role", h.SetRole)

			body, _ := json.Marshal(test.role)

			req := httptest.NewRequest(http.MethodPatch, "/user/"+test.guid.String()+"/role", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)
		})
	}
}
