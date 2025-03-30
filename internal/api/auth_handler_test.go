package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"file-sharing-platform/internal/auth"
	"file-sharing-platform/internal/models"
)

// MockUserRepository is a mock implementation of the UserRepository
type MockUserRepository struct {
	users map[string]models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]models.User),
	}
}

func (r *MockUserRepository) CreateUser(user models.User) (int64, error) {
	user.ID = int64(len(r.users) + 1)
	r.users[user.Email] = user
	return user.ID, nil
}

func (r *MockUserRepository) GetUserByEmail(email string) (models.User, error) {
	if user, exists := r.users[email]; exists {
		return user, nil
	}
	return models.User{}, ErrNotFound
}

var ErrNotFound = errors.New("record not found")

func TestRegisterHandler(t *testing.T) {
	// Create handler with mock repository
	authHandler := &auth.JWTAuth{UserRepo: NewMockUserRepository()}

	// Create test request body
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerReq)

	// Create test request
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	authHandler.Register(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check response body
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	// Check that we got a user ID
	if _, exists := response["user_id"]; !exists {
		t.Errorf("expected user_id in response, got: %v", response)
	}
}
