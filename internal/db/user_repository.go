package db

import (
	"errors"
	"fmt"
	"time"

	"file-sharing-platform/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// VerifyPassword checks if the provided password is correct
func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}

// UserRepository handles user-related database operations
type UserRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(email, password string) (*models.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert the user
	now := time.Now()
	user := models.User{
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO users (email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, created_at, updated_at
	`

	err = r.db.DB.QueryRowx(
		query,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	).StructScan(&user)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Don't return the password hash
	user.Password = ""

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.DB.Get(&user, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.DB.Get(&user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// Authenticate authenticates a user with email and password
func (r *UserRepository) Authenticate(email, password string) (*models.User, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Compare the passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Don't return the password hash
	user.Password = ""

	return user, nil
}
