package password

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

const cost = 12

func Hash(pw string) (string, error) {
	if len(pw) == 0 {return "", errors.New("empty password")}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), cost)
	return string(h), err
}

func Verify(hash, pw string) bool {
	if hash == "" || pw == "" {return false}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
