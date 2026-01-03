package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sanjow-main-api/internal/domain/auth/service"
)

// mockUserService implements service.UserService for testing
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetUserForAuth(ctx context.Context, email string) (*service.AuthUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthUser), args.Error(1)
}

func (m *mockUserService) CheckPassword(passwordHash, password string) bool {
	args := m.Called(passwordHash, password)
	return args.Bool(0)
}

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)
	return router
}

func TestLogin_Success(t *testing.T) {
	// Setup mock
	mockSvc := new(mockUserService)
	user := &service.AuthUser{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
	}
	mockSvc.On("GetUserForAuth", mock.Anything, "test@example.com").Return(user, nil)
	mockSvc.On("CheckPassword", user.PasswordHash, "password123").Return(true)

	// Create service and handler
	authSvc := service.New(mockSvc, "test-secret")
	h := New(authSvc)
	router := setupTestRouter(h)

	// Make request
	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	// Setup mock
	mockSvc := new(mockUserService)
	mockSvc.On("GetUserForAuth", mock.Anything, "wrong@example.com").Return(nil, assert.AnError)

	// Create service and handler
	authSvc := service.New(mockSvc, "test-secret")
	h := New(authSvc)
	router := setupTestRouter(h)

	// Make request
	body := `{"email":"wrong@example.com","password":"badpass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_InvalidRequest(t *testing.T) {
	// Setup mock (won't be called due to validation failure)
	mockSvc := new(mockUserService)

	// Create service and handler
	authSvc := service.New(mockSvc, "test-secret")
	h := New(authSvc)
	router := setupTestRouter(h)

	// Make request with invalid email
	body := `{"email":"not-an-email","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	// Setup mock
	mockSvc := new(mockUserService)
	user := &service.AuthUser{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
	}
	mockSvc.On("GetUserForAuth", mock.Anything, "test@example.com").Return(user, nil)
	mockSvc.On("CheckPassword", user.PasswordHash, "wrongpassword").Return(false)

	// Create service and handler
	authSvc := service.New(mockSvc, "test-secret")
	h := New(authSvc)
	router := setupTestRouter(h)

	// Make request
	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
