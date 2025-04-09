package utils

// importing bcrypt package from crypto to hash passwords easily
import "golang.org/x/crypto/bcrypt"

// hash password while registering
// enter the password eneterd by user
func HashPassword(password string) (string, error) {
	// Generate a bcrypt hash from the password with a cost of 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// check if the hashed password matched the entered password
//during login

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
