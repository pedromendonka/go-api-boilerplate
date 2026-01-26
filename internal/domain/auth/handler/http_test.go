package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"sanjow-nova-api/internal/database/db"
	"sanjow-nova-api/internal/domain/auth/service"
)

// mockUserRepository implements service.UserRepository for testing
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.User), args.Error(1)
}

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)
	return router
}

func TestLogin_Success(t *testing.T) {
	// Hash the test password
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Setup mock
	mockRepo := new(mockUserRepository)
	user := &db.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	// Create service and handler
	authSvc := service.New(mockRepo, "test-secret")
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
	mockRepo := new(mockUserRepository)
	mockRepo.On("GetByEmail", mock.Anything, "wrong@example.com").Return(nil, assert.AnError)

	// Create service and handler
	authSvc := service.New(mockRepo, "test-secret")
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
	mockRepo := new(mockUserRepository)

	// Create service and handler
	authSvc := service.New(mockRepo, "test-secret")
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
	// Hash a different password
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	// Setup mock
	mockRepo := new(mockUserRepository)
	user := &db.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	// Create service and handler
	authSvc := service.New(mockRepo, "test-secret")
	h := New(authSvc)
	router := setupTestRouter(h)

	// Make request with wrong password
	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
