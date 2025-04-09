package utils

import (
	"os"   // used to create secret token
	"time" // time of expiary

	"github.com/golang-jwt/jwt/v4" // package to easily implement jwt
)

var jwtKey = []byte(os.Getenv("JWT_SECRET")) // get secret key from env variable

//HMAC SHA-256 is used for signing.

func GenerateJWT(userID uint) (string, error) { // takes the user id and returns a token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString(jwtKey)
}
